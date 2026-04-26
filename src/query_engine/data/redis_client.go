package data

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/schemas"
	"github.com/redis/go-redis/v9"
)

const (
	SearchTermsKey     = "search_terms"
	SearchTermsDictKey = "search_terms_dict"
)

// RedisClient wraps go-redis for search-term tracking and suggestions.
type RedisClient struct {
	client *redis.Client
	enabled bool
}

// NewRedisClient creates a new instance of RedisClient.
func NewRedisClient(addr, password string, db int) (*RedisClient, error) {
	if addr == "" {
		slog.Warn("Redis disabled: no address provided; suggestions/top-searches tracking will be no-op")
		return &RedisClient{enabled: false}, nil
	}

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
		client:  rdb,
		enabled: true,
	}, nil
}

// NewRedisClientFromURL creates a Redis client using a full redis:// URL.
func NewRedisClientFromURL(rawURL string) (*RedisClient, error) {
	if rawURL == "" {
		slog.Warn("Redis disabled: REDIS_URL is empty; suggestions/top-searches tracking will be no-op")
		return &RedisClient{enabled: false}, nil
	}

	opts, err := redis.ParseURL(rawURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse redis url: %w", err)
	}

	rdb := redis.NewClient(opts)
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	slog.Info("Successfully connected to Redis")
	return &RedisClient{
		client:  rdb,
		enabled: true,
	}, nil
}

// Close closes the Redis connection.
func (r *RedisClient) Close() error {
	if r == nil || !r.enabled || r.client == nil {
		return nil
	}
	return r.client.Close()
}

// IncrSearchTerm increments the search term counter in a sorted set.
func (r *RedisClient) IncrSearchTerm(ctx context.Context, term string) error {
	if r == nil || !r.enabled || r.client == nil {
		return nil
	}
	pipe := r.client.Pipeline()
	pipe.ZIncrBy(ctx, SearchTermsKey, 1, term)
	pipe.ZAdd(ctx, SearchTermsDictKey, redis.Z{Score: 0, Member: term})
	_, err := pipe.Exec(ctx)
	return err
}

// GetTopSearches returns the top N searched terms from Redis.
func (r *RedisClient) GetTopSearches(ctx context.Context, n int) ([]schemas.TermScore, error) {
	if r == nil || !r.enabled || r.client == nil {
		return []schemas.TermScore{}, nil
	}
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
	if r == nil || !r.enabled || r.client == nil {
		return []string{}, nil
	}
	const maxSuggestions = 10
	
	// Efficient prefix search using ZRANGEBYLEX
	op := &redis.ZRangeBy{
		Min:   "[" + prefix,
		Max:   "[" + prefix + "\xff",
		Offset: 0,
		Count: maxSuggestions,
	}
	
	suggestions, err := r.client.ZRangeByLex(ctx, SearchTermsDictKey, op).Result()
	if err != nil && err != redis.Nil {
		return nil, fmt.Errorf("suggestion dict query failed: %w", err)
	}
	
	// Fallback to ZSCAN if dictionary doesn't have it (e.g. legacy data)
	if len(suggestions) == 0 {
		iter := r.client.ZScan(ctx, SearchTermsKey, 0, prefix+"*", 0).Iterator()
		for iter.Next(ctx) {
			member := iter.Val()
			if !iter.Next(ctx) {
				break
			}
			// iter.Val() is now the score, which we ignore
			suggestions = append(suggestions, member)
			if len(suggestions) >= maxSuggestions {
				break
			}
		}
		if err := iter.Err(); err != nil {
			return nil, fmt.Errorf("suggestion scan failed: %w", err)
		}
	}
	
	return suggestions, nil
}
