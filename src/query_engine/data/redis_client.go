package data

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/schemas"
	"github.com/redis/go-redis/v9"
)

const (
	SearchTermsKey = "search_terms"
)

// RedisClient wraps go-redis for search-term tracking and suggestions.
type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new instance of RedisClient.
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("Successfully connected to Redis")

	return &RedisClient{
		client: rdb,
	}, nil
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	return r.client.Close()
}

// IncrSearchTerm increments the search term counter in a sorted set.
func (r *RedisClient) IncrSearchTerm(ctx context.Context, term string) error {
	return r.client.ZIncrBy(ctx, SearchTermsKey, 1, term).Err()
}

// GetTopSearches returns the top N searched terms from Redis.
func (r *RedisClient) GetTopSearches(ctx context.Context, n int) ([]schemas.TermScore, error) {
	results, err := r.client.ZRevRangeWithScores(ctx, SearchTermsKey, 0, int64(n-1)).Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get top searches: %w", err)
	}

	scores := make([]schemas.TermScore, len(results))
	for i, z := range results {
		scores[i] = schemas.TermScore{
			Term:  z.Member.(string),
			Score: z.Score,
		}
	}

	return scores, nil
}

// GetSearchSuggestions returns search suggestions based on a prefix matching.
// It uses ZRANGEBYLEX which is efficient for prefix matching in sorted sets
// where elements have the same score.
func (r *RedisClient) GetSearchSuggestions(ctx context.Context, prefix string) ([]string, error) {
	// For ZRANGEBYLEX to work as a prefix search, all elements should have the same score.
	// However, we are using search_terms which has scores (the counts).
	// If we want to return results based on popularity, we'd need to fetch then filter,
	// or maintain a separate "dictionary" sorted set with score 0.
	
	// For now, let's execute a simpler search or assume we fetch a limited set and match.
	// Actually, ZSCAN with a match pattern is easier but slower for large sets.
	
	// Better approach for autocomplete: Use the search_terms set but filter by prefix.
	// Since we don't have many millions of search terms yet, this is fine.
	
	// Get all members (or a large enough sample) and filter.
	// Optimization: If we had a dedicated "autocomplete" sorted set with 0 scores,
	// we could use ZRANGEBYLEX.
	
	// Let's use ZRANGE with a start/stop if we had 0 scores.
	// Since we have scores, we'll use ZSCAN or just ZREVRANGE and filter.
	
	const maxSuggestions = 10
	var suggestions []string
	
	iter := r.client.ZScan(ctx, SearchTermsKey, 0, prefix+"*", 0).Iterator()
	for iter.Next(ctx) {
		suggestions = append(suggestions, iter.Val())
		if len(suggestions) >= maxSuggestions {
			break
		}
	}
	
	if err := iter.Err(); err != nil {
		return nil, fmt.Errorf("suggestion scan failed: %w", err)
	}
	
	return suggestions, nil
}
