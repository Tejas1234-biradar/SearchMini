package data

import (
	"context"
	"fmt"
	"github.com/redis/go-redis/v9"
	"log"
	"strings"
)

// --------------------- MODEL ---------------------

type Backlinks struct {
	ID    string
	Links map[string]struct{} // set equivalent
}

// --------------------- CLIENT ---------------------

type RedisClient struct {
	client *redis.Client
	ctx    context.Context
}

// Client exposes the underlying Redis client for integration tests.
func (rc *RedisClient) Client() *redis.Client {
	return rc.client
}

func NewRedisClient(host string, port int, password string, db int) (*RedisClient, error) {
	addr := fmt.Sprintf("%s:%d", host, port)

	rdb := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx := context.Background()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	log.Println("Successfully connected to Redis!")

	return &RedisClient{
		client: rdb,
		ctx:    ctx,
	}, nil
}
func (rc *RedisClient) GetAllBacklinksKeys() ([]string, error) {
	if rc.client == nil {
		return nil, fmt.Errorf("redis connection not initialized")
	}

	log.Println("Fetching all backlinks' keys")

	keys, err := rc.client.Keys(rc.ctx, "backlinks:*").Result()
	if err != nil {
		return nil, err
	}

	if len(keys) == 0 {
		log.Println("No backlinks found")
		return nil, nil
	}

	return keys, nil
}
func (rc *RedisClient) GetAllBacklinks(keys []string) ([]Backlinks, error) {
	if rc.client == nil {
		return nil, fmt.Errorf("redis connection not initialized")
	}

	pipe := rc.client.Pipeline()

	cmds := make([]*redis.StringSliceCmd, len(keys))
	urls := make([]string, len(keys))

	for i, key := range keys {
		url := strings.TrimPrefix(key, "backlinks:")
		urls[i] = url
		cmds[i] = pipe.SMembers(rc.ctx, key)
	}

	_, err := pipe.Exec(rc.ctx)
	if err != nil && err != redis.Nil {
		return nil, err
	}

	var result []Backlinks

	for i, cmd := range cmds {
		members, err := cmd.Result()
		if err != nil && err != redis.Nil {
			return nil, err
		}

		// convert []string → set
		linkSet := make(map[string]struct{})
		for _, v := range members {
			linkSet[v] = struct{}{}
		}

		result = append(result, Backlinks{
			ID:    urls[i],
			Links: linkSet,
		})
	}

	return result, nil
}
func (rc *RedisClient) RemoveAllBacklinks(keys []string) (int64, error) {
	if rc.client == nil {
		return 0, fmt.Errorf("redis connection not initialized")
	}

	pipe := rc.client.Pipeline()

	cmds := make([]*redis.IntCmd, len(keys))

	for i, key := range keys {
		cmds[i] = pipe.Del(rc.ctx, key)
	}

	_, err := pipe.Exec(rc.ctx)
	if err != nil {
		return 0, err
	}

	var deleted int64 = 0
	for _, cmd := range cmds {
		val, _ := cmd.Result()
		deleted += val
	}

	if deleted == 0 {
		return 0, nil
	}

	return deleted, nil
}
