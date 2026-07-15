package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/hotel-voice-agent/gateway/internal/cache"
	"github.com/hotel-voice-agent/gateway/internal/db"
	grpcserver "github.com/hotel-voice-agent/gateway/internal/grpc"
)

func main() {
	log.Println("Starting Hotel Voice Agent Gateway...")

	// 1. Initialize SQLite Database
	if err := db.InitDB("hotel.db"); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.DB.Close()

	repo := db.NewSQLiteRepository(db.DB)

	// 2. Initialize Redis Cache
	// Assuming local Redis on default port 6379
	if err := cache.InitRedis("localhost:6379"); err != nil {
		log.Printf("Warning: Redis initialization failed, caching may not work: %v", err)
	} else {
		defer cache.RDB.Close()
	}

	// 3. Start gRPC Server in a goroutine
	go func() {
		if err := grpcserver.StartServer(50051, repo); err != nil {
			log.Fatalf("Failed to start gRPC server: %v", err)
		}
	}()

	// 4. Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Gateway gracefully...")
}
