package event_logger

import (
	"boralabs/internal/chain"
	"boralabs/pkg/datastore/mongodb"
	boraLabsErr "boralabs/pkg/error"
	"boralabs/pkg/util"
	"fmt"
	"github.com/ethereum/go-ethereum/core/types"
	"log"
)

type Logger struct {
	*chain.Contract
}

const retryLimit = 3

func NewLogger(contract *chain.Contract, evtName string) *Logger {
	var err error
	evt := chain.Event{Cont: contract}
	if evt, err = evt.New(evtName); err != nil {
		panic(err)
	}

	return &Logger{Contract: contract}
}

func (l *Logger) Collect(evtName string) {
	var err error
	evt := chain.Event{Cont: l.Contract}
	if evt, err = evt.New(evtName); err != nil {
		panic(err)
	}

	var logs []types.Log
	startBlock := mongodb.CalcFromBlock(collectionName)
	tryCnt := 1
	log.Printf("[%s] Starting events collector :: %s :: [Start BlockNumber - %d]\n", util.NowInKst().String(), evtName, startBlock)
	defer log.Printf("[%s] End events collector :: %s :: [Start BlockNumber - %d]\n", util.NowInKst().String(), evtName, startBlock)
	for tryCnt <= retryLimit {
		logs, err = chain.GovCont.FilterLogs(evt.Signature, startBlock)
		if err != nil {
			log.Println(err)
		}
		tryCnt++
	}

	if len(logs) == 0 {
		log.Printf("not found logs :: %s\n", evtName)
		return
	}

	for _, eLog := range logs {
		if err = chain.GovCont.UnpackLogData(evt.Out, evt.Name, eLog); err != nil {
			log.Println(fmt.Sprintf(boraLabsErr.FailedParseLogData, err))
		}
		// log save
		err = evt.SaveLog("", eLog)
		if err != nil {
			log.Println(fmt.Sprintf(boraLabsErr.FailedSaveLogData, err))
		}
	}
}
