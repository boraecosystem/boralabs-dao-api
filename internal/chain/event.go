package chain

import (
	"boralabs/internal/model"
	"boralabs/pkg/datastore/mongodb"
	boraLabsErr "boralabs/pkg/error"
	"boralabs/pkg/util"
	"context"
	"errors"
	"fmt"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/iancoleman/strcase"
	"go.mongodb.org/mongo-driver/bson"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"math/big"
	"time"
)

type Event struct {
	Name      string
	Signature string
	Out       any
	Cont      *Contract
}

const (
	NameProposals            = "proposals"
	NameVotes                = "votes"
	EventNameProposalCreated = "ProposalCreated"
	EventNameVoteCast        = "VoteCast"
	ProposalStatePending     = "pending"
	ProposalStateActive      = "active"
	ProposalStateClosed      = "closed"
)

func (e *Event) SaveLog(proposalID string, log types.Log) error {
	evt := *e
	var ok bool
	coll := mongodb.DB.Collection(fmt.Sprintf("%s_logs", strcase.ToSnake(e.Name)))
	block, err := e.Cont.BlockByNumber(big.NewInt(int64(log.BlockNumber)))
	if err != nil {
		return errors.New(fmt.Sprintf("%s\n%v", boraLabsErr.FailedBlockByNumber, err))
	}

	switch evt.Name {
	case EventNameProposalCreated:
		var data *model.ProposalCreatedLog
		if data, ok = evt.Out.(*model.ProposalCreatedLog); !ok {
			util.ErrorLog(errors.New(boraLabsErr.FailedParseLogData))
		}
		if proposalID != "" && data.ProposalId.String() != proposalID {
			return nil
		}

		m := model.MongoProposalCreatedLog{
			ProposalId:     data.ProposalId.String(),
			Proposer:       data.Proposer.String(),
			Targets:        util.ConvArrayToStringArr(data.Targets),
			Values:         util.ConvArrayToStringArr(data.Values),
			Signatures:     data.Signatures,
			Calldatas:      util.ConvArrayToStringArr(data.Calldatas),
			VoteStart:      time.Unix(data.VoteStart.Int64(), 0),
			VoteEnd:        time.Unix(data.VoteEnd.Int64(), 0),
			Description:    data.Description,
			Log:            log,
			BlockCreatedAt: time.Unix(int64(block.Time()), 0),
			UpdatedAt:      time.Now(),
		}

		// log update
		opt := options.Update().SetUpsert(true)
		_, err = coll.UpdateOne(context.Background(), bson.D{
			{Key: "proposal_id", Value: m.ProposalId},
		}, bson.D{{"$set", m}}, opt)
		if err != nil {
			util.ErrorLog(errors.New(fmt.Sprintf(boraLabsErr.FailedSaveLogData, err)))
			return err
		}

		// get proposal for totalSupply
		var proposal model.Proposal
		err = mongodb.DB.Collection(NameProposals).FindOne(context.Background(), bson.D{
			{"proposal_id", data.ProposalId.String()},
		}).Decode(&proposal)
		if err != nil {
			if errors.Is(err, mongo2.ErrNoDocuments) {
				return nil
			}
			return errors.New(fmt.Sprintf("Error find proposal %s\n%v", data.ProposalId.String(), err))
		}

		// get total supply
		totalSupply, totalVotingPower, votingRatio := CalcTotalSupply(proposal, proposal.StartDate)

		// proposals update
		proposalState := CalcProposalState(m.VoteStart, m.VoteEnd)
		_, err = mongodb.DB.Collection(NameProposals).UpdateOne(
			context.Background(),
			bson.D{{Key: "proposal_id", Value: m.ProposalId}},
			bson.D{
				{"$set", bson.D{
					{Key: "block_number", Value: block.Header().Number.Uint64()},
					{Key: "tx_hash", Value: log.TxHash.Hex()},
					{Key: "start_date", Value: m.VoteStart},
					{Key: "end_date", Value: m.VoteEnd},
					{Key: "proposer", Value: m.Proposer},
					{Key: "total_voting_power", Value: totalVotingPower.String()},
					{Key: "total_supply", Value: totalSupply.String()},
					{Key: "voting_ratio", Value: votingRatio.String()},
					{Key: "state", Value: proposalState},
				}},
			}, opt)
		if err != nil {
			util.ErrorLog(errors.New(fmt.Sprintf(boraLabsErr.FailedSaveLogData, err)))
			return err
		}
		util.Log(fmt.Sprintf("Successfully %s [%s] [%s]", evt.Name, data.ProposalId.String(), proposalState))

	case EventNameVoteCast:
		var data *model.VoteCastLog
		if data, ok = e.Out.(*model.VoteCastLog); !ok {
			return errors.New(boraLabsErr.FailedParseLogData)
		}

		if proposalID != "" && data.ProposalId.String() != proposalID {
			return nil
		}

		v := model.MongoVoteCastLog{
			Voter:      data.Voter.String(),
			ProposalId: data.ProposalId.String(),
			Support:    data.Support,
			Weight:     uint8(data.Weight.Uint64()),
			Reason:     data.Reason,
			CreatedAt:  data.CreatedAt,
		}
		// log update
		opt := options.Update().SetUpsert(true)
		_, err = coll.UpdateOne(context.Background(), bson.D{
			{Key: "proposal_id", Value: v.ProposalId},
			{Key: "voter", Value: v.Voter},
		}, bson.D{{"$set", v}}, opt)
		if err != nil {
			return errors.New(fmt.Sprintf(boraLabsErr.FailedSaveLogData, err))
		}

		if checkExistsProposal(data.ProposalId.String()) == false {
			return errors.New(fmt.Sprintf("Not found proposal %s", data.ProposalId.String()))
		}

		_, err = mongodb.DB.Collection(NameVotes).
			UpdateOne(context.Background(), bson.D{
				{Key: "proposal_id", Value: v.ProposalId},
				{Key: "wallet_address", Value: v.Voter},
			},
				bson.D{
					{"$setOnInsert", bson.D{
						{Key: "id", Value: mongodb.NextSequence(NameVotes)},
						{Key: "proposal_id", Value: v.ProposalId},
						{Key: "wallet_address", Value: v.Voter},
						{Key: "voting_power", Value: data.Weight.String()},
						{Key: "status", Value: data.Support},
						{Key: "tx_hash", Value: log.TxHash.Hex()},
						{Key: "created_at", Value: time.Now()},
					}},
				}, opt)
		if err != nil {
			util.ErrorLog(errors.New(fmt.Sprintf(boraLabsErr.FailedSaveLogData, err)))
		}
		util.Log(fmt.Sprintf("Successfully %s [%s]", evt.Name, data.ProposalId.String()))

		if err = updateTotalSupply(data.ProposalId.String()); err != nil {
			util.Log(fmt.Sprintf("Failed update total supply :: [Proposal ID:%s]", data.ProposalId.String()))
		}
	}

	return nil
}

func checkExistsProposal(proposalId string) bool {
	// is existing check
	coll := mongodb.DB.Collection(NameProposals)
	res := coll.FindOne(context.Background(), bson.D{
		{Key: "proposal_id", Value: proposalId},
	})
	if res.Err() == nil {
		return true
	}
	return false
}

func (e *Event) New(name string) (evt Event, err error) {
	var event abi.Event
	var has bool
	if event, has = e.Cont.events[name]; !has {
		err = errors.New(fmt.Sprintf(boraLabsErr.InvalidEventName, name))
		return
	}

	e.Name = name
	e.Signature = event.Sig
	switch name {
	case EventNameProposalCreated:
		e.Out = &model.ProposalCreatedLog{}
	case EventNameVoteCast:
		e.Out = &model.VoteCastLog{}
	default:
		err = errors.New(fmt.Sprintf(boraLabsErr.InvalidEventName, name))
	}
	evt = *e
	return
}

// CalcProposalState Calculate the proposal state based on the current time.
func CalcProposalState(startDt, endDt time.Time) (state string) {
	now := time.Now()

	if util.IsDebug() {
		log.Printf("Debug :: Now - %s / Start/End Time - %s\n", now.String(), startDt.String()+"/"+endDt.String())
	}
	if now.Location().String() != "UTC" {
		now = now.UTC()
	}
	// If the time zone is not UTC, forcibly change all to be based on UTC.
	times := &[]time.Time{now, startDt, endDt}
	for _, t := range *times {
		if t.Location().String() != "UTC" {
			t = t.UTC()
		}
	}

	state = ProposalStatePending
	// If the start date is equal to or in the past of the current time, and the end date is in the future, then it is considered active
	if startDt.Unix() <= now.Unix() && endDt.Unix() > now.Unix() {
		state = ProposalStateActive
	} else if endDt.Unix() <= now.Unix() {
		state = ProposalStateClosed
	}
	return
}

func updateTotalSupply(proposalId string) error {
	var err error
	var proposal model.Proposal
	err = mongodb.DB.Collection(NameProposals).FindOne(context.Background(), bson.D{
		{"proposal_id", proposalId},
	}).Decode(&proposal)
	if err != nil {
		if errors.Is(err, mongo2.ErrNoDocuments) {
			return errors.New(fmt.Sprintf("Not found proposal %s", proposalId))
		}
		return errors.New(fmt.Sprintf("Error find proposal %s\n%v", proposalId, err))
	}

	// get total supply
	totalSupply, totalVotingPower, votingRatio := CalcTotalSupply(proposal, proposal.StartDate)

	// proposals update
	if totalSupply.Cmp(big.NewInt(0)) > 0 || totalVotingPower.Cmp(big.NewInt(0)) > 0 || votingRatio.Cmp(big.NewInt(0)) > 0 {
		opt := options.Update().SetUpsert(true)
		_, err = mongodb.DB.Collection(NameProposals).UpdateOne(
			context.Background(),
			bson.D{{Key: "proposal_id", Value: proposalId}},
			bson.D{
				{"$set", bson.D{
					{Key: "total_voting_power", Value: totalVotingPower.String()},
					{Key: "total_supply", Value: totalSupply.String()},
					{Key: "voting_ratio", Value: votingRatio.String()},
				}},
			}, opt)
		if err != nil {
			util.ErrorLog(errors.New(fmt.Sprintf(boraLabsErr.FailedSaveLogData, err)))
			return err
		}
	}
	return err
}

func CalcTotalSupply(proposal model.Proposal, startDt time.Time) (totalSupply, totalVotingPower, votingRatio *big.Int) {
	totalSupply = big.NewInt(0)
	totalVotingPower = big.NewInt(0)
	votingRatio = big.NewInt(0)

	if proposal.State == ProposalStatePending || proposal.State == "" || startDt.IsZero() { // When in pending status, it's not possible to query the total supply from the contract.
		log.Printf("Skipping total supply query: Proposal is in pending state (%s), state is empty (%v), or start date is zero (%v)", proposal.State, proposal.State == "", startDt.IsZero())
		return
	}

	// get total supply
	totalSupplyList, _ := GetPastTotalSupply(startDt)

	if len(totalSupplyList) > 0 && totalSupplyList[0] != nil {
		totalSupply = totalSupplyList[0].(*big.Int)
	} else {
		totalSupply, _ = big.NewInt(0).SetString(proposal.TotalSupply, 10)
	}

	if totalSupply == nil || totalSupply.Cmp(big.NewInt(0)) == 0 {
		util.ErrorLog(errors.New(fmt.Sprintf("[%s | %s] get totalSupply failed :: [%s] :: DB [%s] / BlockChain [%s]\n", proposal.ProposalID, proposal.State, startDt.String(), proposal.TotalSupply, totalSupply)))
		return
	}

	// calculate voting status
	if totalSupply.String() != "" {
		voteColl := mongodb.DB.Collection(NameVotes)
		cursor, err := voteColl.Find(context.Background(), bson.D{
			{"proposal_id", proposal.ProposalID},
			{"voting_power", bson.M{"$ne": "0"}},
		})
		if err != nil {
			panic(err)
		}

		defer cursor.Close(context.TODO())

		for cursor.Next(context.TODO()) {
			var vote model.VoteCast
			err = cursor.Decode(&vote)
			if err != nil {
				panic(err)
			}

			tmpVotingPower, ok := big.NewInt(0).SetString(vote.VotingPower, 10)
			if !ok {
				log.Println("convert failed voting power :: ", vote.VotingPower)
			} else {
				totalVotingPower = totalVotingPower.Add(totalVotingPower, tmpVotingPower)
			}
		}
		// calculate voting ratio
		votingRatio = votingRatio.Mul(totalVotingPower, big.NewInt(int64(100)))
		votingRatio = votingRatio.Div(votingRatio, totalSupply)
	}
	return
}
