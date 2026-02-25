package database

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/utils"
	"github.com/redis/go-redis/v9"
)

// Redis Setup
type Database struct {
	Client  *redis.Client
	Context context.Context
} //context is used to add timeouts and thrad safe operations wtver that means
func (db *Database) ConnectToRedis(redisHost, redisPort, redisPassword, redisDB string) error {
	log.Println("Initializing Redis")
	dbIndex, err := strconv.Atoi(redisDB)
	if err != nil {
		return fmt.Errorf("Could Parse DB value: %v\n", err)
	}
	db.Client = redis.NewClient(&redis.Options{
		Addr:     redisHost + ":" + redisPort,
		Password: redisPassword,
		DB:       dbIndex,
	})
	db.Context = context.Background()
	_, err = db.Client.Ping(db.Context).Result()
	if err != nil {
		return fmt.Errorf("Couldnt connect to RedisClient maybe give up? [%v,%v]:%v", redisHost, redisPassword, err)

	}
	log.Println("Connected to redis")
	return nil
}
func (db *Database) PushURL(rawURL string, score float64) error {
	rawURL, err := utils.StripURL(rawURL)
	if err != nil {
		return fmt.Errorf("Could not strip the url: %w;", err)
	}
	normalizedUrl, err := utils.NormalizeURL(rawURL)
	if err != nil {
		return fmt.Errorf("Could not normalize the url: %w", err)
	}
	err = db.Client.ZAdd(db.Context, utils.CrawlerQueueKey, redis.Z{
		Score:  score,
		Member: normalizedUrl,
	}).Err()
	if err != nil {
		return fmt.Errorf("Could not add url to queue: %w", err)
	}
	fmt.Printf("Pushed %v (%v) to queue\n", rawURL, normalizedUrl)
	return nil
}
func (db *Database) ExistsInQueue(rawUrl string) (float64, bool) {
	normalizedURL, err := utils.NormalizeURL(rawUrl)
	if err != nil {
		return 0.0, false
	}
	result, err := db.Client.ZScore(db.Context, utils.CrawlerQueueKey, normalizedURL).Result()
	if err != nil {
		return 0.0, false
	}
	return result, true
}
func (db *Database) HasURLBeenVisited(normalizedURL string) (bool, error) {
	return false, nil //FIXME:Temporary Fix
}
func (db *Database) VisitPage(NormalizedUrl string) error {
	return nil //FIXME:Temporary Fix
}
func (db *Database) PopURL() (string, float64, string, error) {
	result, err := db.Client.BZPopMin(db.Context, utils.Timeout, utils.CrawlerQueueKey).Result()
	if err != nil {
		return "", 0.0, "", fmt.Errorf("Could not pop from queue: %w", err)
	}
	normalizedUrl := result.Z.Member.(string)
	raw_url := fmt.Sprintf("https://%v", normalizedUrl)
	return raw_url, result.Z.Score, normalizedUrl, nil
}

func (db *Database) PopSignalQueue() (string, error) {
	result, err := db.Client.BRPop(db.Context, 0, utils.SignalQueueKey).Result()
	if err != nil {
		return "", fmt.Errorf("Could not pop from signal queue: %v\n", err)
	}

	return result[1], nil
}

func (db *Database) GetIndexerQueueSize() (int64, error) {
	size, err := db.Client.LLen(db.Context, utils.IndexerQueueKey).Result()
	if err != nil {
		return -1, fmt.Errorf("Could not get %v size: %v\n", utils.IndexerQueueKey, err)
	}

	return size, nil
}
