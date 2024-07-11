package mongodb

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"math"
)

type Paginator struct {
	Page      int64 `json:"page"`
	Limit     int64 `json:"limit"`
	Total     int64 `json:"total"`
	TotalPage int64 `json:"totalPage"`
}

func NewPaginator() (p *Paginator) {
	p = &Paginator{}
	p.Page = 1
	p.Limit = 20
	p.TotalPage = 1
	return p
}

func (p *Paginator) Calculate(collName string, filter interface{}, sort bson.D) (*mongo.Cursor, error) {
	var err error
	ctx := context.Background()
	coll := DB.Collection(collName)

	// calculate total count
	p.Total, err = coll.CountDocuments(context.Background(), filter)
	if err != nil {
		return nil, err
	}

	// when total exceeds limit
	if p.Total > p.Limit {
		p.TotalPage = int64(math.Ceil(float64(p.Total) / float64(p.Limit)))
	}

	skip := (p.Page * p.Limit) - p.Limit
	fOpt := options.FindOptions{Limit: &p.Limit, Skip: &skip}
	fOpt.SetSort(sort)

	return coll.Find(ctx, filter, &fOpt)
}
