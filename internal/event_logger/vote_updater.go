package event_logger

import (
	"boralabs/internal/chain"
	"boralabs/internal/model"
	"boralabs/pkg/datastore/mongodb"
	boraLabsErr "boralabs/pkg/error"
	"boralabs/pkg/notification"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
	"time"
)

type VoteUpdater struct {
	*model.VoteCast
}

const collectionNameVote = "votes"

func (p *VoteUpdater) Default() {
	p.ID = mongodb.NextSequence(collectionNameVote)
	p.CreatedAt = time.Now()
}

func (p *VoteUpdater) Update(startBlock uint64) {
	defer func() {
		if r := recover(); r != nil {
			notification.SendAll(fmt.Sprintf("Panic Vote AppendSave %v", r))
		}
	}()

	event := &chain.Event{Cont: chain.GovCont}
	evt, err := event.New(chain.EventNameVoteCast)
	if err != nil {
		panic(err)
	}

	var logs []types.Log
	from := mongodb.CalcFromBlock(collectionNameVote)
	if startBlock != 0 {
		from = startBlock
	}
	logs, err = chain.GovCont.FilterLogs(evt.Signature, from)
	if err != nil {
		panic(err)
	}

	defaultTerm := 1 * time.Second
	tryCnt := 1
	errCnt := 0
	for tryCnt <= retryLimit {
		for _, eLog := range logs {
			if err = chain.GovCont.UnpackLogData(evt.Out, evt.Name, eLog); err != nil {
				log.Println(fmt.Sprintf(boraLabsErr.FailedParseLogData, err))
			}
			err = evt.SaveLog(p.ProposalId, eLog)
			if err != nil {
				errCnt++
				panic(err)
			}
		}
		tryCnt++
		time.Sleep(defaultTerm)
		defaultTerm = defaultTerm + defaultTerm // 1 - 2 - 4
	}
}
