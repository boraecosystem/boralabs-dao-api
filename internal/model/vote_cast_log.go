package model

import (
	"github.com/ethereum/go-ethereum/common"
	"math/big"
	"time"
)

// VoteCastLog VoteCast Log Data
type VoteCastLog struct {
	Voter      common.Address `bson:"voter,omitempty"`
	ProposalId *big.Int       `bson:"proposal_id,omitempty"`
	Support    uint8          `bson:"support,omitempty"`
	Weight     *big.Int       `bson:"weight,omitempty"` // voting power
	Reason     string         `bson:"reason,omitempty"`
	CreatedAt  time.Time      `bson:"created_at"`
}

type MongoVoteCastLog struct {
	Voter      string    `bson:"voter,omitempty"`
	ProposalId string    `bson:"proposal_id,omitempty"`
	Support    uint8     `bson:"support,omitempty"`
	Weight     uint8     `bson:"weight,omitempty"` // voting power
	Reason     string    `bson:"reason,omitempty"`
	CreatedAt  time.Time `bson:"created_at"`
}
