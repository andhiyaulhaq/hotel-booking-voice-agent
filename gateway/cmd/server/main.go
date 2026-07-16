package main

import (
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/hotel-voice-agent/gateway/internal/cache"
	"github.com/hotel-voice-agent/gateway/internal/db"
	grpcserver "github.com/hotel-voice-agent/gateway/internal/grpc"
	"github.com/hotel-voice-agent/gateway/internal/payments"
	"github.com/hotel-voice-agent/gateway/internal/ws"
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

	// 4. Initialize Xendit
	payments.InitXendit()

	// 5. Start HTTP Server for WebSockets and Webhooks
	http.HandleFunc("/ws", ws.HandleConnections)
	http.HandleFunc("/webhooks/xendit", payments.HandleWebhook(repo))
	
	go func() {
		log.Println("HTTP Server listening on :8080 (WebSockets & Webhooks)")
		if err := http.ListenAndServe(":8080", nil); err != nil {
			log.Fatalf("HTTP server failed: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down Gateway gracefully...")
}
