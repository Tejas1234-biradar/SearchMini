package data

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

type mongoTestEnv struct {
	ctx         context.Context
	mongoClient *MongoClient
	rawMongo    *mongo.Database
	cleanup     func()
}

func setupMongoTestEnv(t *testing.T) *mongoTestEnv {
	t.Helper()
	ctx := context.Background()

	// ---- Read connection details from environment ----
	mongoHost := getEnv("MONGO_HOST", "localhost")
	mongoPortStr := getEnv("MONGO_PORT", "27017")
	mongoPort, _ := strconv.Atoi(mongoPortStr)
	mongoUser := getEnv("MONGO_USERNAME", "root")
	mongoPass := getEnv("MONGO_PASSWORD", "password")
	mongoDB := getEnv("MONGO_DB", "testdb")

	// MongoClient for query engine
	mongoClient, err := NewMongoClient(mongoHost, mongoUser, mongoPass, mongoDB, mongoPort)
	if err != nil {
		t.Fatalf("Failed to create MongoClient: %v", err)
	}

	// Raw mongo for setup/assertions
	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s?authSource=admin", mongoUser, mongoPass, mongoHost, mongoPort, mongoDB)
	mClient, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		t.Fatalf("Failed to connect to raw mongo: %v", err)
	}
	db := mClient.Database(mongoDB)

	// Verify connection
	if err := mClient.Ping(ctx, nil); err != nil {
		t.Fatalf("Failed to ping mongo: %v", err)
	}

	cleanup := func() {
		// Clean up collections used in tests to avoid pollution
		_ = db.Collection(wordsCollection).Drop(ctx)
		_ = db.Collection(metadataCollection).Drop(ctx)
		_ = db.Collection(outlinksCollection).Drop(ctx)
		_ = db.Collection(backlinksCollection).Drop(ctx)
		_ = db.Collection(pageRankCollection).Drop(ctx)

		_ = mongoClient.Close(ctx)
		_ = mClient.Disconnect(ctx)
	}

	return &mongoTestEnv{
		ctx:         ctx,
		mongoClient: mongoClient,
		rawMongo:    db,
		cleanup:     cleanup,
	}
}

func TestMongoClient_GetStats(t *testing.T) {
	env := setupMongoTestEnv(t)
	defer env.cleanup()

	// 1. Initially empty (after drop in cleanup/setup)
	count, err := env.mongoClient.GetStats(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(0), count)

	// 2. Insert metadata
	_, err = env.rawMongo.Collection(metadataCollection).InsertOne(env.ctx, bson.M{"_id": "http://example.com", "title": "Example"})
	require.NoError(t, err)

	count, err = env.mongoClient.GetStats(env.ctx)
	require.NoError(t, err)
	assert.Equal(t, int64(1), count)
}

func TestMongoClient_GetMetadataBatch(t *testing.T) {
	env := setupMongoTestEnv(t)
	defer env.cleanup()

	urls := []string{"http://a.com", "http://b.com"}
	for _, url := range urls {
		_, err := env.rawMongo.Collection(metadataCollection).InsertOne(env.ctx, bson.M{
			"_id":   url,
			"title": "Title for " + url,
		})
		require.NoError(t, err)
	}

	result, err := env.mongoClient.GetMetadataBatch(env.ctx, urls)
	require.NoError(t, err)
	assert.Len(t, result, 2)
	assert.Equal(t, "Title for http://a.com", result["http://a.com"].Title)
}

func TestMongoClient_GetPageConnections(t *testing.T) {
	env := setupMongoTestEnv(t)
	defer env.cleanup()

	url := "http://main.com"
	outlink := "http://out.com"
	backlink := "http://back.com"

	// Seed metadata for link enrichment
	_, _ = env.rawMongo.Collection(metadataCollection).InsertMany(env.ctx, []interface{}{
		bson.M{"_id": outlink, "title": "Out Title"},
		bson.M{"_id": backlink, "title": "Back Title"},
	})

	// Seed outlinks and backlinks
	_, _ = env.rawMongo.Collection(outlinksCollection).InsertOne(env.ctx, bson.M{"_id": url, "links": []string{outlink}})
	_, _ = env.rawMongo.Collection(backlinksCollection).InsertOne(env.ctx, bson.M{"_id": url, "links": []string{backlink}})

	outs, backs, err := env.mongoClient.GetPageConnections(env.ctx, url)
	require.NoError(t, err)

	require.Len(t, outs, 1)
	assert.Equal(t, outlink, outs[0].URL)
	assert.Equal(t, "Out Title", outs[0].Title)

	require.Len(t, backs, 1)
	assert.Equal(t, backlink, backs[0].URL)
	assert.Equal(t, "Back Title", backs[0].Title)
}

func TestMongoClient_SearchPages(t *testing.T) {
	env := setupMongoTestEnv(t)
	defer env.cleanup()

	// Seed data
	// URL A: word1 (tfidf 10), word2 (tfidf 5)
	// URL B: word1 (tfidf 8)
	// PageRank: URL A (1.0), URL B (2.0)
	
	_, _ = env.rawMongo.Collection(wordsCollection).InsertMany(env.ctx, []interface{}{
		bson.M{"word": "word1", "url": "http://a.com", "tfidf": 10.0},
		bson.M{"word": "word2", "url": "http://a.com", "tfidf": 5.0},
		bson.M{"word": "word1", "url": "http://b.com", "tfidf": 8.0},
	})

	_, _ = env.rawMongo.Collection(metadataCollection).InsertMany(env.ctx, []interface{}{
		bson.M{"_id": "http://a.com", "title": "Title A"},
		bson.M{"_id": "http://b.com", "title": "Title B"},
	})

	_, _ = env.rawMongo.Collection(pageRankCollection).InsertMany(env.ctx, []interface{}{
		bson.M{"_id": "http://a.com", "score": 1.0},
		bson.M{"_id": "http://b.com", "score": 2.0},
	})

	// Search for ["word1"]
	// http://a.com -> tfidf 10, pr 1 -> score = 0.6*10 + 0.4*1 = 6.4
	// http://b.com -> tfidf 8, pr 2 -> score = 0.6*8 + 0.4*2 = 5.6
	results, total, err := env.mongoClient.SearchPages(env.ctx, []string{"word1"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	require.Len(t, results, 2)
	assert.Equal(t, "http://a.com", results[0].URL)
	assert.Equal(t, 6.4, results[0].Score)
	assert.Equal(t, "http://b.com", results[1].URL)
	assert.Equal(t, 5.6, results[1].Score)

	// Search for ["word1", "word2"]
	// http://a.com -> matchCount 2, cumWeight 15, pr 1 -> score = 0.6*15 + 0.4*1 = 9.4
	// http://b.com -> matchCount 1, cumWeight 8, pr 2 -> score = 0.6*8 + 0.4*2 = 5.6
	results, total, err = env.mongoClient.SearchPages(env.ctx, []string{"word1", "word2"}, 1, 10)
	require.NoError(t, err)
	assert.Equal(t, 2, total)
	assert.Equal(t, "http://a.com", results[0].URL)
	assert.Equal(t, 9.4, results[0].Score)
}

func TestMongoClient_GetTopRankedPages(t *testing.T) {
	env := setupMongoTestEnv(t)
	defer env.cleanup()

	_, _ = env.rawMongo.Collection(wordsCollection).InsertMany(env.ctx, []interface{}{
		bson.M{"word": "w1", "url": "http://a.com", "tfidf": 100.0},
		bson.M{"word": "w2", "url": "http://b.com", "tfidf": 50.0},
		bson.M{"word": "w3", "url": "http://b.com", "tfidf": 60.0}, // total B = 110
	})

	results, err := env.mongoClient.GetTopRankedPages(env.ctx, 10)
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, "http://b.com", results[0].URL)
	assert.Equal(t, 110.0, results[0].Score)
	assert.Equal(t, "http://a.com", results[1].URL)
	assert.Equal(t, 100.0, results[1].Score)
}
