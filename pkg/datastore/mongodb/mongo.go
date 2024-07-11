package mongodb

import (
	"boralabs/config"
	"boralabs/internal/model"
	"context"
	"errors"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var (
	Conn               *mongo.Client
	DB                 *mongo.Database
	latestCollectBlock uint64
)

func New() {
	var err error
	credential := options.Credential{
		Username: config.C.GetString("mongo.user"),
		Password: config.C.GetString("mongo.pass"),
	}
	clientOptions := options.Client().
		ApplyURI(fmt.Sprintf("mongodb://%s:%s", config.C.GetString("mongo.host"), config.C.GetString("mongo.port"))).
		SetAuth(credential)
	Conn, err = mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		log.Fatal(err)
	}

	// Check the connection
	err = Conn.Ping(context.TODO(), nil)

	if err != nil {
		log.Fatal(err)
	}
	DB = Conn.Database("boralabs")
	createCollections([]string{
		"proposals",
		"proposal_created_logs",
		"vote_cast_logs",
	})
	log.Println("MongoDB Connected")
}

func createCollections(collections []string) {
	for _, collection := range collections {
		if err := DB.CreateCollection(context.Background(), collection, nil); err != nil {
			var e mongo.CommandError
			if errors.As(err, &e) && e.Code == 48 {
				log.Printf("Exists Collection :: %s %v\n", collection, err)
				continue
			} else {
				log.Printf("Failed CreateCollection :: %s %v\n", collection, err)
			}
		}
	}
}

func IsDuplicateErr(err error, isPassDup bool) error {
	if isDup(err) == true && isPassDup == true {
		return nil
	}
	return err
}

func isDup(err error) bool {
	var e mongo.WriteException
	if errors.As(err, &e) {
		for _, we := range e.WriteErrors {
			if we.Code == 11000 {
				return true
			}
		}
	}
	return false
}

func NextSequence(coll string) uint64 {
	var lastDoc bson.M
	opts := options.FindOne().SetSort(bson.M{"id": -1})
	_ = DB.Collection(coll).FindOne(context.Background(), bson.M{}, opts).Decode(&lastDoc)
	if lastDoc == nil {
		return 1
	}
	lastId := lastDoc["id"].(int64)

	return uint64(lastId) + 1
}

func CalcFromBlock(collName string) (from uint64) {
	from = config.C.GetUint64("fromBlock")
	isRewind := config.C.GetBool("isRewind")
	opt := options.Find().SetSort(bson.D{{"block_number", 1}})
	find, err := DB.Collection(collName).Find(context.Background(),
		bson.D{
			{"state", bson.M{"$ne": "closed"}}, // exclude close state
		}, opt)
	if err != nil {
		return from
	}

	if from > 0 && isRewind {
		return from // If the isRewind environment variable or config value is set, it will be configured to the initial block number.
	}

	var proposals []model.Proposal
	err = find.All(context.Background(), &proposals)
	if err != nil || len(proposals) == 0 {
		if latestCollectBlock > 0 {
			from = latestCollectBlock
		}
		return from
	}

	if proposals[0].BlockNumber > 0 {
		from = proposals[0].BlockNumber
		latestCollectBlock = from
	}
	if from == 0 {
		return from
	}

	return from - 1
}
