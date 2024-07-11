package model

import (
	"time"
)

const (
	StatusNo = iota
	StatusYes
	StatusAbstain
)

type VoteCast struct {
	ID            uint64    `bson:"id" json:"id"`
	ProposalId    string    `bson:"proposal_id" json:"proposal_id,omitempty"`
	WalletAddress string    `bson:"wallet_address" json:"walletAddress"`
	VotingPower   string    `bson:"voting_power" json:"votingPower"` // voting power
	Weight        uint8     `bson:"weight" json:"-"`                 // Percentage of DAO tokens owned by the user compared to the total DAO tokens.
	Status        uint8     `bson:"status" json:"status"`
	TxHash        string    `bson:"tx_hash" json:"txhash"`
	CreatedAt     time.Time `bson:"created_at" json:"created_at"`
}
