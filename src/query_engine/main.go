package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/data"
	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/pagerank"
	"github.com/Tejas1234-biradar/DBMS-CP/src/query_engine/routes"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func main() {
	port := getEnv("PORT", "8080")

	// DB connection settings
	mongoURI := getEnv("MONGO_URI", "")
	mongoHost := getEnv("MONGO_HOST", "localhost")
	mongoPort, _ := strconv.Atoi(getEnv("MONGO_PORT", "27017"))
	mongoUser := getEnv("MONGO_USERNAME", "admin")
	mongoPass := getEnv("MONGO_PASSWORD", "pass123")
	mongoDB := getEnv("MONGO_DB", "mongo-test")

	redisURL := getEnv("REDIS_URL", "")
	redisHost := getEnv("REDIS_HOST", "")
	redisPort := getEnv("REDIS_PORT", "6379")
	redisPass := getEnv("REDIS_PASSWORD", "")
	redisDB, _ := strconv.Atoi(getEnv("REDIS_DB", "0"))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Connect to Mongo
	var mongoClient *data.MongoClient
	var err error
	if mongoURI != "" {
		slog.Info("Connecting to MongoDB using MONGO_URI", "db", mongoDB)
		mongoClient, err = data.NewMongoClientFromURI(mongoURI, mongoDB)
	} else {
		slog.Info("Connecting to MongoDB", "host", mongoHost, "port", mongoPort, "db", mongoDB, "user", mongoUser)
		mongoClient, err = data.NewMongoClient(mongoHost, mongoUser, mongoPass, mongoDB, mongoPort)
	}
	if err != nil {
		log.Fatalf("Failed to connect to Mongo: %v", err)
	}
	defer mongoClient.Close(ctx)

	// Compute PageRank on startup so query scoring always has fresh graph scores.
	slog.Info("Computing PageRank scores")
	if err := pagerank.Run(ctx, mongoClient.Database(), pagerank.DefaultConfig()); err != nil {
		log.Fatalf("Failed to compute PageRank: %v", err)
	}
	slog.Info("PageRank scores updated")

	// Connect to Redis
	var redisClient *data.RedisClient
	if redisURL != "" {
		slog.Info("Connecting to Redis using REDIS_URL")
		redisClient, err = data.NewRedisClientFromURL(redisURL)
	} else {
		addr := ""
		if redisHost != "" {
			addr = redisHost + ":" + redisPort
		}
		slog.Info("Connecting to Redis", "host", redisHost, "port", redisPort, "db", redisDB)
		redisClient, err = data.NewRedisClient(addr, redisPass, redisDB)
	}
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redisClient.Close()

	// Initialize router
	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	// Register routes
	routes.Register(r, mongoClient, redisClient)

	// Graceful shutdown
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: r,
	}

	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		slog.Warn("Shutting down server...")
		
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer shutdownCancel()
		
		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("HTTP shutdown error: %v", err)
		}
		cancel()
	}()

	slog.Info("Query engine listening", "port", port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatal(err)
	}
}
