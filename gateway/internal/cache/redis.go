package cache

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/hotel-voice-agent/gateway/internal/db"
	"github.com/redis/go-redis/v9"
)

var RDB *redis.Client
var ctx = context.Background()

func InitRedis(addr string) error {
	RDB = redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	_, err := RDB.Ping(ctx).Result()
	if err != nil {
		return fmt.Errorf("failed to ping redis: %w", err)
	}

	log.Println("Redis initialized successfully.")
	return nil
}

// GetAvailableRooms implements the Cache-Aside pattern
func GetAvailableRooms(repo db.BookingRepository, roomType string) (int, error) {
	key := fmt.Sprintf("hotel:availability:%s", roomType)
	
	// 1. Try Cache
	val, err := RDB.Get(ctx, key).Int()
	if err == nil {
		// Cache Hit
		return val, nil
	} else if err != redis.Nil {
		// Real error
		log.Printf("Redis error: %v", err)
		// Fallback to DB
	}

	// 2. Cache Miss - Query DB
	available, err := repo.GetAvailableRooms(roomType)
	if err != nil {
		return 0, err
	}

	// 3. Populate Cache with TTL of 15 minutes
	err = RDB.Set(ctx, key, available, 15*time.Minute).Err()
	if err != nil {
		log.Printf("Failed to set redis cache: %v", err)
	}

	return available, nil
}

// InvalidateAvailability clears the cache for a room type (e.g. after a booking)
func InvalidateAvailability(roomType string) {
	key := fmt.Sprintf("hotel:availability:%s", roomType)
	err := RDB.Del(ctx, key).Err()
	if err != nil {
		log.Printf("Failed to delete redis cache for %s: %v", key, err)
	}
}
