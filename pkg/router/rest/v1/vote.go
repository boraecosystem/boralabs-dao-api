package v1

import (
	"boralabs/internal/event_logger"
	"boralabs/internal/model"
	mongoDb "boralabs/pkg/datastore/mongodb"
	"boralabs/pkg/router/rest"
	"context"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"log"
)

const (
	voteCollectionName = "votes"
)

type VoteV1 struct {
	rest.Response
}

type requestStatus struct {
	Status uint8 `form:"status" json:"status"`
	Page   int64 `form:"page" json:"page"`
}

var statusString = map[uint8]string{}

func init() {
	statusString[model.StatusNo] = "no"
	statusString[model.StatusYes] = "yes"
	statusString[model.StatusAbstain] = "abstain"
}

func (v VoteV1) findAll(c *gin.Context) {
	v.Context = c
	req := requestStatus{Page: 1}
	if err := c.BindQuery(&req); err != nil {
		v.JsonError(err)
		return
	}

	var filter bson.D
	filter = append(filter, bson.E{Key: "proposal_id", Value: c.Param("id")})

	if c.Request.URL.Query().Get("status") != "" {
		filter = append(filter, bson.E{Key: "status", Value: req.Status})
	}

	v.BaseResponse.Paginator = mongoDb.NewPaginator()
	v.BaseResponse.Paginator.Limit = 30
	v.BaseResponse.Paginator.Page = req.Page

	cursor, err := v.BaseResponse.Paginator.Calculate(voteCollectionName, filter, bson.D{{"id", -1}})
	if err != nil {
		v.JsonError(err)
		return
	}

	var votes []model.VoteCast
	var vote model.VoteCast
	for cursor.Next(context.Background()) {
		if err = cursor.Decode(&vote); err != nil {
			log.Println(err)
		}

		votes = append(votes, vote)
	}

	v.BaseResponse.Data = gin.H{
		"items": votes,
	}
	v.BaseResponse.IsPaging = true
	v.Json()
}

func (v VoteV1) create(c *gin.Context) {
	v.Context = c
	var req model.VoteCast
	// validate request to structure
	if err := c.Bind(&req); err != nil {
		v.Code = 400
		v.JsonError(err)
		return
	}
	req.ProposalId = c.Param("id")

	// is exists proposal
	res := mongoDb.DB.Collection(mongoCollectionName).FindOne(context.Background(), bson.D{
		{Key: "proposal_id", Value: req.ProposalId},
	})
	if res.Err() != nil {
		v.Code = 400
		v.JsonError(res.Err())
		return
	}

	// db update
	var proposal model.Proposal
	err := mongoDb.DB.Collection(mongoCollectionName).FindOne(context.Background(), bson.D{
		{"proposal_id", req.ProposalId},
	}).Decode(&proposal)

	blockNumber := uint64(0)
	lProposal, err := latestProposal(proposal.ID)
	if err == nil {
		blockNumber = lProposal.BlockNumber
	}

	updater := event_logger.VoteUpdater{VoteCast: &req}
	updater.Update(blockNumber)

	v.Code = 200
	v.BaseResponse.Data = gin.H{
		"votes": []model.VoteCast{req},
	}
	v.Json()
	return
}
