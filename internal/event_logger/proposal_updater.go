package event_logger

import (
	"boralabs/internal/chain"
	"boralabs/internal/model"
	"boralabs/pkg/datastore/mongodb"
	boraLabsErr "boralabs/pkg/error"
	"boralabs/pkg/util"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"time"
)

type ProposalUpdater struct {
	*model.Proposal
}

const collectionName = "proposals"

func (p *ProposalUpdater) Default() {
	p.ID = mongodb.NextSequence(collectionName)
	p.CreatedAt = time.Now()
	p.State = chain.ProposalStatePending
}

func (p *ProposalUpdater) Update(startBlock uint64) (ret bool) {
	defer func() {
		if r := recover(); r != nil {
			util.ErrorLog(errors.New(fmt.Sprintf("Panic Proposal AppendSave %v", r)))
			util.PrintStackTrace()
		}
	}()

	event := &chain.Event{Cont: chain.GovCont}
	evt, err := event.New(chain.EventNameProposalCreated)
	if err != nil {
		panic(err)
	}

	var logs []types.Log
	from := mongodb.CalcFromBlock(collectionName)
	if startBlock != 0 {
		from = startBlock
	}
	defaultTerm := 1 * time.Second
	tryCnt := 1
	var data *model.ProposalCreatedLog
	var ok bool
	matchCnt := 0
	for tryCnt <= retryLimit {
		logs, err = chain.GovCont.FilterLogs(evt.Signature, from)
		if err != nil {
			panic(err)
		}

		for _, eLog := range logs {
			if data, ok = evt.Out.(*model.ProposalCreatedLog); !ok {
				util.ErrorLog(errors.New(boraLabsErr.FailedParseLogData))
				continue
			}

			if err = chain.GovCont.UnpackLogData(evt.Out, evt.Name, eLog); err != nil {
				log.Println(fmt.Sprintf(boraLabsErr.FailedParseLogData, err))
			}
			err = evt.SaveLog(data.ProposalId.String(), eLog)
			if err != nil {
				log.Println(err)
				continue
			}

			if data.ProposalId.String() == p.ProposalID {
				matchCnt++
			}
		}

		if matchCnt > 0 {
			ret = true
			break
		}

		time.Sleep(defaultTerm)
		defaultTerm = defaultTerm + defaultTerm // 1 - 2 - 4
		tryCnt++
	}
	return
}
