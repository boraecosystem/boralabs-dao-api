package v1

import (
	"boralabs/config"
	"boralabs/internal/chain"
	"boralabs/internal/event_logger"
	"boralabs/internal/model"
	mongoDb "boralabs/pkg/datastore/mongodb"
	boraLabsErr "boralabs/pkg/error"
	"boralabs/pkg/notification"
	"boralabs/pkg/router/rest"
	"boralabs/pkg/util"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	mongo2 "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"io"
	"log"
	"net/http"
	"time"
)

const (
	mongoCollectionName = "proposals"
	filterMaskTypeAll   = "all"
)

type ProposalV1 struct {
	rest.Response
}

type ProposalV1FindAll struct {
	State string `form:"state" json:"state"`
	Page  int64  `form:"page" json:"page"`
}

type ProposalLatest struct {
	NextId uint64 `json:"next_id"`
}

type FilterMaskReq struct {
	filterType string
	sentence   string
}

type FilterMaskRes struct {
	Success bool        `json:"success"`
	Error   interface{} `json:"error"`
	Payload struct {
		FilterWordCount int      `json:"filterWordCount"`
		FilterWords     []string `json:"filterWords"`
		OriginSentence  string   `json:"originSentence"`
		MaskingSentence string   `json:"maskingSentence"`
	} `json:"payload"`
}

func (p ProposalV1) routes(group *gin.RouterGroup) {
	group = group.Group("proposals")
	{
		group.POST("", p.create)
		group.GET("", p.findAll)
		group.GET(":id", p.find)
		vote := group.Group(":id/votes")
		{
			voteV1 := VoteV1{}
			vote.POST("", voteV1.create) // Create a vote
			vote.GET("", voteV1.findAll) // Vote Information
		}
		group.GET("latest-id", p.findLatestId)
	}
}

/*
*
Create a proposal
*/
func (p ProposalV1) create(c *gin.Context) {
	p.Context = c
	var req model.Proposal
	// validate request to structure
	if err := c.Bind(&req); err != nil {
		p.Code = http.StatusBadRequest
		p.JsonError(err)
		return
	}

	coll := mongoDb.DB.Collection(mongoCollectionName)
	// is existing check
	res := coll.FindOne(context.Background(), bson.D{
		{Key: "proposal_id", Value: req.ProposalID},
	})
	if res.Err() == nil {
		p.Code = http.StatusConflict
		p.BaseResponse.Message = boraLabsErr.FailedExistsProposal
		p.JsonError(res.Err())
		return
	}

	// Filter banned words API information is used to mask the proposal title if it contains restricted words.
	if config.C.GetString("filter_api.host") != "" {
		filterReq := FilterMaskReq{
			filterType: filterMaskTypeAll,
			sentence:   req.Title,
		}
		filterReqBytes, err := json.Marshal(filterReq)
		if err != nil {
			p.Code = http.StatusInternalServerError
			p.JsonError(res.Err())
			return
		}

		filterRes, err := http.Post(fmt.Sprintf("http://%s:%s/filter/v1/mask", config.C.GetString("filter_api.host"), config.C.GetString("filter_api.port")), "application/json", bytes.NewBuffer(filterReqBytes))
		if err != nil {
			p.JsonError(res.Err())
			return
		}
		resBody, err := io.ReadAll(filterRes.Body)
		if err != nil {
			p.JsonError(res.Err())
			return
		}
		var filterMaskRes FilterMaskRes
		if err = json.Unmarshal(resBody, &filterMaskRes); err != nil {
			p.JsonError(res.Err())
			return
		}
		if filterMaskRes.Success {
			req.Title = filterMaskRes.Payload.MaskingSentence
		}
	}

	// calculate sequence
	updater := event_logger.ProposalUpdater{Proposal: &req}
	updater.Default()

	// insert document
	_, err := coll.InsertOne(context.Background(), req)
	if mongoDb.IsDuplicateErr(err, false) != nil {
		p.JsonError(err)
		return
	}

	blockNumber := uint64(0)
	lProposal, err := latestProposal(req.ID)
	if err == nil {
		blockNumber = lProposal.BlockNumber
	}

	// db update
	updateStartTime := time.Now() // Record start time
	updateRes := updater.Update(blockNumber)
	if updateRes != true {
		p.Code = http.StatusInternalServerError
		p.BaseResponse.Message = boraLabsErr.FailedUpdateLogData
		p.JsonError(errors.New("error retrieving blockchain event data for proposal"))
		return
	}
	elapsedTime := time.Since(updateStartTime) // Calculate elapsed time
	log.Printf("updater.Update() execution time: %s", elapsedTime.String())

	// response
	p.Code = http.StatusCreated
	p.BaseResponse.Data = gin.H{
		"proposals": []model.Proposal{req},
	}
	p.Json()
	return
}

func latestProposal(ID uint64) (result model.Proposal, err error) {
	// Set search criteria and sorting criteria
	filter := bson.M{
		"id":           bson.M{"$lt": ID}, // Condition where the value is less than the value in column id
		"block_number": bson.M{"$ne": 0},  // Condition where the value in column block_number is not zero
	}
	findOptions := options.FindOne().SetSort(bson.D{{"id", -1}}) // Sort in descending order based on column A

	// Single document search
	coll := mongoDb.DB.Collection(mongoCollectionName)
	err = coll.FindOne(context.TODO(), filter, findOptions).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo2.ErrNoDocuments) {
			log.Println("No document found")
		}
	}
	return
}

func (p ProposalV1) find(c *gin.Context) {
	p.Context = c
	proposalId := c.Param("id")
	coll := mongoDb.DB.Collection(mongoCollectionName)

	var proposal model.Proposal
	err := coll.FindOne(context.Background(), bson.D{
		{"proposal_id", proposalId},
	}).Decode(&proposal)
	if err != nil {
		p.Code = http.StatusInternalServerError
		if errors.Is(err, mongo2.ErrNoDocuments) {
			p.Code = http.StatusNotFound
			return
		}
		p.JsonError(err)
		return
	}

	// update proposal state
	go updateState([]model.Proposal{proposal})

	p.BaseResponse.Data = proposal
	p.Json()
}

func (p ProposalV1) findAll(c *gin.Context) {
	p.Context = c
	req := ProposalV1FindAll{Page: 1}
	if err := c.BindQuery(&req); err != nil {
		p.JsonError(err)
		return
	}

	filter := bson.D{{}}
	if req.State != "" {
		filter = bson.D{{Key: "state", Value: req.State}}
	}
	p.BaseResponse.Paginator = mongoDb.NewPaginator()
	p.BaseResponse.Paginator.Limit = 8
	p.BaseResponse.Paginator.Page = req.Page

	cursor, err := p.BaseResponse.Paginator.Calculate(mongoCollectionName, filter, bson.D{{"id", -1}})
	if err != nil {
		p.JsonError(err)
		return
	}

	var proposals []model.Proposal
	if err = cursor.All(context.Background(), &proposals); err != nil {
		p.JsonError(err)
		return
	}

	// update proposal state
	go updateState(proposals)

	p.BaseResponse.Data = proposals
	p.BaseResponse.IsPaging = true
	p.Json()
}

func (p ProposalV1) findLatestId(c *gin.Context) {
	result := ProposalLatest{
		NextId: mongoDb.NextSequence(mongoCollectionName),
	}
	p.Context = c
	p.BaseResponse.Data = result
	p.Json()
}

func updateState(p []model.Proposal) {
	var lastProposalId uint64
	defer func() {
		if err := recover(); err != nil {
			notification.SendAll(fmt.Sprintf("Panic Proposal [%d] AppendSave %v", lastProposalId, err))
		}
	}()
	// proposals update
	log.Println("=== Start update Proposals State ===")
	for _, proposal := range p {
		lastProposalId = proposal.ID
		_, err := mongoDb.DB.Collection(mongoCollectionName).UpdateOne(
			context.Background(),
			bson.D{{Key: "proposal_id", Value: proposal.ProposalID}},
			bson.D{
				{"$set", bson.D{
					{Key: "state", Value: chain.CalcProposalState(proposal.StartDate, proposal.EndDate)},
				}},
			}, nil)
		if err != nil {
			util.ErrorLog(errors.New(fmt.Sprintf(boraLabsErr.FailedSaveLogData, err)))
		}
		log.Printf("=== Updated Proposals State [%s] ===\n", proposal.ProposalID)
	}
	log.Println("=== End update Proposals State ===")
}
