
package main

import (
	"context"
	"testing"
	"time"

	"github.com/Tejas1234-biradar/DBMS-CP/src/backlinks-processor/data"
)

func TestPipelineFlow(t *testing.T) {
	ctx := context.Background()

	// ---- INIT CLIENTS ----
	redisClient, err := data.NewRedisClient("localhost", 6379, "", 0)
	if err != nil {
		t.Fatalf("Redis init failed: %v", err)
	}

	mongoClient, err := data.NewMongoClient("localhost", "", "", "test", 27017)
	if err != nil {
		t.Fatalf("Mongo init failed: %v", err)
	}

	// ---- SEED REDIS ----
	key := "backlinks:test.com"

	err = redisClient.Client().SAdd(ctx, key, "a.com", "b.com", "c.com").Err()
	if err != nil {
		t.Fatalf("Failed seeding redis: %v", err)
	}

	// ---- RUN LOGIC ----
	keys, _ := redisClient.GetAllBacklinksKeys()
	backlinks, _ := redisClient.GetAllBacklinks(keys)

	if len(backlinks) == 0 {
		t.Fatalf("Expected backlinks, got none")
	}

	_, err = mongoClient.SaveAllBacklinks(ctx, backlinks)
	if err != nil {
		t.Fatalf("Mongo save failed: %v", err)
	}

	_, err = redisClient.RemoveAllBacklinks(keys)
	if err != nil {
		t.Fatalf("Redis delete failed: %v", err)
	}

	// ---- VERIFY MONGO ----
	time.Sleep(1 * time.Second)

	collection := mongoClient.DB().Collection("backlinks")

	count, err := collection.CountDocuments(ctx, map[string]interface{}{
		"_id": "test.com",
	})

	if err != nil {
		t.Fatalf("Mongo query failed: %v", err)
	}

	if count == 0 {
		t.Fatalf("Expected document in Mongo, found none")
	}
}