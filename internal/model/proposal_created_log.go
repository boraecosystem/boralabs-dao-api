package model

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"math/big"
	"time"
)

// ProposalCreatedLog ProposalCreated Log Data
// event ProposalCreated(uint256 proposalId, address proposer, address[] targets, uint256[] values, string[] signatures, bytes[] calldatas, uint256 voteStart, uint256 voteEnd, string description)
type ProposalCreatedLog struct {
	ProposalId     *big.Int         `bson:"proposal_id" json:"proposal_id,omitempty"`
	Proposer       common.Address   `bson:"proposer" json:"proposer,omitempty"`
	Targets        []common.Address `bson:"targets" json:"targets,omitempty"`
	Values         []*big.Int       `bson:"values" json:"values,omitempty"`
	Signatures     []string         `bson:"signatures" json:"signatures,omitempty"`
	Calldatas      []common.Hash    `bson:"call_data" json:"calldatas,omitempty"`
	VoteStart      *big.Int         `bson:"vote_start" json:"vote_start,omitempty"`
	VoteEnd        *big.Int         `bson:"vote_end" json:"vote_end,omitempty"`
	Description    string           `bson:"description" json:"description,omitempty"`
	Log            types.Log        `bson:"log"`
	BlockCreatedAt time.Time        `bson:"block_created_at" json:"block_created_at"`
	CreatedAt      time.Time        `bson:"created_at" json:"created_at"`
}

type MongoProposalCreatedLog struct {
	ProposalId     string    `bson:"proposal_id" json:"proposal_id,omitempty"`
	Proposer       string    `bson:"proposer" json:"proposer,omitempty"`
	Targets        []string  `bson:"targets" json:"targets,omitempty"`
	Values         []string  `bson:"values" json:"values,omitempty"`
	Signatures     []string  `bson:"signatures" json:"signatures,omitempty"`
	Calldatas      []string  `bson:"call_data" json:"calldatas,omitempty"`
	VoteStart      time.Time `bson:"vote_start" json:"vote_start,omitempty"`
	VoteEnd        time.Time `bson:"vote_end" json:"vote_end,omitempty"`
	Description    string    `bson:"description" json:"description,omitempty"`
	Log            types.Log `bson:"log"`
	BlockCreatedAt time.Time `bson:"block_created_at" json:"block_created_at"`
	UpdatedAt      time.Time `bson:"updated_at" json:"updated_at"`
}
