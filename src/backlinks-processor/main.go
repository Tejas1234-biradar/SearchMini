package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Tejas1234-biradar/DBMS-CP/src/backlinks-processor/data"
)

func getEnv(key, defaultVal string) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultVal
	}
	return val
}

func main() {
	// --------------------- CONTEXT + SHUTDOWN ---------------------
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Termination signal received: %v - shutting down...\n", sig)
		cancel()
	}()

	// --------------------- ENV ---------------------

	// REDIS
	redisHost := getEnv("REDIS_HOST", "localhost")
	redisPort, _ := strconv.Atoi(getEnv("REDIS_PORT", "6379"))
	redisPassword := os.Getenv("REDIS_PASSWORD")
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	// MONGO
	mongoHost := getEnv("MONGO_HOST", "localhost")
	mongoPort, _ := strconv.Atoi(getEnv("MONGO_PORT", "27017"))
	mongoPassword := os.Getenv("MONGO_PASSWORD")
	mongoDB := getEnv("MONGO_DB", "test")
	mongoUsername := os.Getenv("MONGO_USERNAME")

	// --------------------- INIT REDIS ---------------------
	log.Println("Initializing Redis...")

	redisClient, err := data.NewRedisClient(redisHost, redisPort, redisPassword, redisDB)
	if err != nil {
		log.Println("Could not initialize Redis:", err)
		os.Exit(1)
	}

	// --------------------- INIT MONGO ---------------------
	log.Println("Initializing Mongo...")

	mongoClient, err := data.NewMongoClient(mongoHost, mongoUsername, mongoPassword, mongoDB, mongoPort)
	if err != nil {
		log.Println("Could not initialize Mongo:", err)
		os.Exit(1)
	}

	// --------------------- LOOP ---------------------
	for {
		select {
		case <-ctx.Done():
			log.Println("Service stopped.")
			return
		default:
		}

		log.Println("Processing backlinks...")

		// 1. Get keys
		keys, err := redisClient.GetAllBacklinksKeys()
		if err != nil || len(keys) == 0 {
			log.Println("No backlinks to process - sleeping...")
			sleepWithContext(ctx, 10*time.Second)
			continue
		}

		// 2. Get backlinks
		backlinks, err := redisClient.GetAllBacklinks(keys)
		if err != nil {
			log.Println("Could not fetch backlinks - retry:", err)
			continue
		}

		// 3. Delete from Redis
		log.Println("Removing backlinks from Redis...")
		deleted, err := redisClient.RemoveAllBacklinks(keys)
		if err != nil {
			log.Println("Failed deleting backlinks:", err)
		} else if deleted > 0 {
			log.Printf("%d backlinks removed from Redis!\n", deleted)
		}

		// 4. Save to Mongo
		_, err = mongoClient.SaveAllBacklinks(ctx, backlinks)
		if err != nil {
			log.Println("Failed saving to Mongo:", err)
			continue
		}

		sleepWithContext(ctx, 10*time.Second)
	}
}

// graceful sleep
func sleepWithContext(ctx context.Context, d time.Duration) {
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	timeout := time.After(d)

	for {
		select {
		case <-ctx.Done():
			return
		case <-timeout:
			return
		case <-ticker.C:
		}
	}
}
