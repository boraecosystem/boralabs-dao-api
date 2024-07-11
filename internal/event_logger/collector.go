package event_logger

import (
	"boralabs/internal/chain"
	"log"
)

type Collector struct {
}

func (c Collector) Collect() {
	eventsMap := make(map[string]string)
	eventsMap[chain.EventNameProposalCreated] = "proposals"
	eventsMap[chain.EventNameVoteCast] = "vote_casts"
	log.Println("Starting events collector")
	defer log.Println("End events collector")
	for evtName := range eventsMap {
		evtLogger := NewLogger(chain.GovCont, evtName)
		evtLogger.Collect(evtName)
	}
}
