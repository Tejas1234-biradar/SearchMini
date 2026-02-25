package database

import (
	"testing"

	"github.com/Tejas1234-biradar/DBMS-CP/src/crawler/core/utils"
)

func setupTestDB(t *testing.T) *Database {
	db := &Database{}
	err := db.ConnectToRedis("localhost", "6379", "", "0")
	if err != nil {
		t.Fatalf("Failed to connect to Redis: %v", err)
	}

	// Clean up test keys
	db.Client.Del(db.Context, utils.CrawlerQueueKey)
	db.Client.Del(db.Context, utils.SignalQueueKey)
	db.Client.Del(db.Context, utils.IndexerQueueKey)

	return db
}

func TestPushAndExists(t *testing.T) {
	db := setupTestDB(t)

	testURL := "https://example.com/page1"
	score := 10.0

	err := db.PushURL(testURL, score)
	if err != nil {
		t.Fatalf("PushURL failed: %v", err)
	}

	s, ok := db.ExistsInQueue(testURL)
	if !ok {
		t.Fatalf("ExistsInQueue returned false for pushed URL")
	}

	if s != score {
		t.Fatalf("Expected score %.2f, got %.2f", score, s)
	}

	t.Logf("Push and Exists passed: %s (score %.2f)", testURL, s)
}

func TestPopURL(t *testing.T) {
	db := setupTestDB(t)

	testURL := "https://example.com/page2"
	score := 5.0

	err := db.PushURL(testURL, score)
	if err != nil {
		t.Fatalf("PushURL failed: %v", err)
	}

	rawURL, poppedScore, normalizedURL, err := db.PopURL()
	if err != nil {
		t.Fatalf("PopURL failed: %v", err)
	}

	t.Logf("Popped: rawURL=%s, normalizedURL=%s, score=%.2f", rawURL, normalizedURL, poppedScore)

	expectedNormalized, _ := utils.NormalizeURL(testURL)
	if normalizedURL != expectedNormalized {
		t.Fatalf("Normalized URL mismatch: got %s, want %s", normalizedURL, expectedNormalized)
	}
}

func TestPopSignalQueue(t *testing.T) {
	db := setupTestDB(t)

	// Push a signal manually
	signal := "crawl_done"
	db.Client.LPush(db.Context, utils.SignalQueueKey, signal)

	result, err := db.PopSignalQueue()
	if err != nil {
		t.Fatalf("PopSignalQueue failed: %v", err)
	}

	if result != signal {
		t.Fatalf("Signal mismatch: got %s, want %s", result, signal)
	}

	t.Logf("PopSignalQueue passed: %s", result)
}

func TestGetIndexerQueueSize(t *testing.T) {
	db := setupTestDB(t)

	// Push a dummy item
	db.Client.RPush(db.Context, utils.IndexerQueueKey, "dummy_page")

	size, err := db.GetIndexerQueueSize()
	if err != nil {
		t.Fatalf("GetIndexerQueueSize failed: %v", err)
	}

	if size != 1 {
		t.Fatalf("Expected queue size 1, got %d", size)
	}

	t.Logf("GetIndexerQueueSize passed: %d", size)
}
