package data

import (
	"context"
	"fmt"
	"log"
	"net/url"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// --------------------- MODEL ---------------------



// --------------------- CLIENT ---------------------

type MongoClient struct {
	client *mongo.Client
	db     *mongo.Database
}

// DB exposes the underlying Mongo database for integration tests.
func (mc *MongoClient) DB() *mongo.Database {
	return mc.db
}

// Constructor
func NewMongoClient(host, username, password, dbName string, port int) (*MongoClient, error) {
	encodedPassword := url.QueryEscape(password)

	uri := fmt.Sprintf(
		"mongodb://%s:%s@%s:%d/%s?authSource=admin&authMechanism=SCRAM-SHA-256",
		username,
		encodedPassword,
		host,
		port,
		dbName,
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to mongo: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return nil, fmt.Errorf("failed to ping mongo: %w", err)
	}

	log.Println("Successfully connected to MongoDB!")

	return &MongoClient{
		client: client,
		db:     client.Database(dbName),
	}, nil
}

// --------------------- BULK OPS ---------------------

func (mc *MongoClient) PerformBatchOperations(ctx context.Context, operations []mongo.WriteModel, collectionName string) (*mongo.BulkWriteResult, error) {
	if mc.client == nil {
		return nil, fmt.Errorf("mongo connection not initialized")
	}

	if len(operations) == 0 {
		return nil, fmt.Errorf("no operations to perform")
	}

	collection := mc.db.Collection(collectionName)

	res, err := collection.BulkWrite(ctx, operations)
	if err != nil {
		return nil, err
	}

	return res, nil
}

// --------------------- BACKLINKS ---------------------

func (mc *MongoClient) SaveAllBacklinks(ctx context.Context, backlinks []Backlinks) (*mongo.BulkWriteResult, error) {
	if mc.client == nil {
		return nil, fmt.Errorf("mongo connection not initialized")
	}

	var operations []mongo.WriteModel

	for _, backlinkData := range backlinks {
		for _, link := range backlinkData.Links {
			op := mongo.NewUpdateOneModel().
				SetFilter(bson.M{"_id": backlinkData.ID}).
				SetUpdate(bson.M{
					"$addToSet": bson.M{
						"links": link,
					},
				}).
				SetUpsert(true)

			operations = append(operations, op)
		}
	}

	return mc.PerformBatchOperations(ctx, operations, "backlinks")
}
