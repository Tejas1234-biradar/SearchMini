package data

import (
	"context"
	"os"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupRedisTestEnv(t *testing.T) (*RedisClient, context.Context, func()) {
	t.Helper()
	ctx := context.Background()

	redisHost := os.Getenv("REDIS_HOST")
	if redisHost == "" {
		redisHost = "localhost"
	}
	redisPort := os.Getenv("REDIS_PORT")
	if redisPort == "" {
		redisPort = "6379"
	}
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDBStr := os.Getenv("REDIS_DB")
	redisDB := 0
	if redisDBStr != "" {
		redisDB, _ = strconv.Atoi(redisDBStr)
	}

	addr := redisHost + ":" + redisPort
	client, err := NewRedisClient(addr, redisPassword, redisDB)
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	cleanup := func() {
		// Clean up the test key
		_ = client.client.Del(ctx, SearchTermsKey).Err()
		_ = client.Close()
	}

	return client, ctx, cleanup
}

func TestRedisClient_IncrSearchTerm(t *testing.T) {
	client, ctx, cleanup := setupRedisTestEnv(t)
	defer cleanup()

	term := "golang"
	err := client.IncrSearchTerm(ctx, term)
	require.NoError(t, err)

	score, err := client.client.ZScore(ctx, SearchTermsKey, term).Result()
	require.NoError(t, err)
	assert.Equal(t, 1.0, score)

	// Increment again
	err = client.IncrSearchTerm(ctx, term)
	require.NoError(t, err)

	score, err = client.client.ZScore(ctx, SearchTermsKey, term).Result()
	require.NoError(t, err)
	assert.Equal(t, 2.0, score)
}

func TestRedisClient_GetTopSearches(t *testing.T) {
	client, ctx, cleanup := setupRedisTestEnv(t)
	defer cleanup()

	terms := map[string]float64{
		"apple":  10,
		"banana": 20,
		"cherry": 15,
	}

	for term, count := range terms {
		for i := 0; i < int(count); i++ {
			_ = client.IncrSearchTerm(ctx, term)
		}
	}

	top, err := client.GetTopSearches(ctx, 2)
	require.NoError(t, err)
	require.Len(t, top, 2)

	assert.Equal(t, "banana", top[0].Term)
	assert.Equal(t, 20.0, top[0].Score)
	assert.Equal(t, "cherry", top[1].Term)
	assert.Equal(t, 15.0, top[1].Score)
}

func TestRedisClient_GetSearchSuggestions(t *testing.T) {
	client, ctx, cleanup := setupRedisTestEnv(t)
	defer cleanup()

	terms := []string{"go", "golang", "google", "apple", "goroutine"}
	for _, term := range terms {
		_ = client.IncrSearchTerm(ctx, term)
	}

	suggestions, err := client.GetSearchSuggestions(ctx, "go")
	require.NoError(t, err)
	
	// We expect "go", "golang", "google", "goroutine"
	assert.Contains(t, suggestions, "go")
	assert.Contains(t, suggestions, "golang")
	assert.Contains(t, suggestions, "google")
	assert.Contains(t, suggestions, "goroutine")
	assert.NotContains(t, suggestions, "apple")
}
