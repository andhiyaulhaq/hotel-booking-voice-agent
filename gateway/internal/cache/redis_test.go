package cache

import (
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/hotel-voice-agent/gateway/internal/db"
)

func TestCacheAsidePattern(t *testing.T) {
	// 1. Setup Miniredis
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("Failed to start miniredis: %v", err)
	}
	defer mr.Close()

	// Initialize the real Redis client pointing to miniredis
	err = InitRedis(mr.Addr())
	if err != nil {
		t.Fatalf("Failed to init redis client: %v", err)
	}
	defer RDB.Close()

	// 2. Setup In-Memory SQLite Repository for fallback
	err = db.InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init db: %v", err)
	}
	defer db.DB.Close()
	repo := db.NewSQLiteRepository(db.DB)

	roomType := "standard"
	key := "hotel:availability:" + roomType

	// Step 1: Initial call should be a Cache Miss and fetch from DB (10 available)
	available, err := GetAvailableRooms(repo, roomType)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if available != 10 {
		t.Fatalf("Expected 10 available, got %d", available)
	}

	// Verify it was set in Redis
	val, err := mr.Get(key)
	if err != nil {
		t.Fatalf("Expected key to be set in redis, got error: %v", err)
	}
	if val != "10" {
		t.Fatalf("Expected redis value 10, got %s", val)
	}

	// Step 2: Manually change Redis value to simulate Cache Hit
	mr.Set(key, "99")

	available2, err := GetAvailableRooms(repo, roomType)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if available2 != 99 {
		t.Fatalf("Expected cache hit value 99, got %d", available2)
	}

	// Step 3: Test InvalidateAvailability
	InvalidateAvailability(roomType)
	if mr.Exists(key) {
		t.Fatalf("Expected key to be deleted from redis")
	}

	// Step 4: After invalidation, it should fall back to DB again (10 available)
	available3, err := GetAvailableRooms(repo, roomType)
	if err != nil {
		t.Fatalf("Expected no error, got: %v", err)
	}
	if available3 != 10 {
		t.Fatalf("Expected DB fallback value 10, got %d", available3)
	}
}
