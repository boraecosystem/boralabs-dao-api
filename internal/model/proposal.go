package model

import (
	"time"
)

type Proposal struct {
	ID               uint64    `bson:"id" json:"id"`
	Title            string    `bson:"title" json:"title"`
	Description      string    `bson:"description" json:"description,omitempty"`
	Target           []string  `bson:"target" json:"targets"`
	Value            []string  `bson:"value" json:"values"`
	CallData         []string  `bson:"call_data" json:"calldatas"`
	TxHash           string    `bson:"tx_hash" json:"tx_hash,omitempty"`
	ScenarioType     uint8     `bson:"scenario_type" json:"scenario_type"`
	TotalSupply      string    `bson:"total_supply" json:"total_supply"`
	TotalVotingPower string    `bson:"total_voting_power" json:"total_voting_power"`
	VotingRatio      string    `bson:"voting_ratio" json:"voting_ratio"`
	StartDate        time.Time `bson:"start_date" json:"start_date,omitempty"`
	EndDate          time.Time `bson:"end_date" json:"end_date,omitempty"`
	Proposer         string    `bson:"proposer" json:"proposer,omitempty"`
	State            string    `bson:"state" json:"state"`
	BlockNumber      uint64    `bson:"block_number" json:"block_number,omitempty"`
	ProposalID       string    `bson:"proposal_id" json:"proposal_id" binding:"required"`
	CreatedAt        time.Time `bson:"created_at" json:"-"`
}
