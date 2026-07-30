package main

import (
	"encoding/json"
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
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found or error loading it. Using environment variables.")
	}

	log.Println("Starting Hotel Voice Agent Gateway...")

	// 1. Initialize PostgreSQL Database
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		dbURL = "postgres://hotel_user:hotel_password@localhost:5432/hotel_db?sslmode=disable"
	}
	if err := db.InitDB(dbURL); err != nil {
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

	// 5. Initialize Python Brain Client
	agentClient, err := grpcserver.NewAgentClient("localhost:50052")
	if err != nil {
		log.Fatalf("Failed to initialize Agent Client: %v", err)
	}
	defer agentClient.Close()

	// 6. Initialize WebSocket Handler
	wsHandler := ws.NewWebSocketHandler(agentClient)

	// 7. Start HTTP Server for WebSockets, Webhooks, and API
	http.HandleFunc("/ws", wsHandler.HandleConnections)
	http.HandleFunc("/webhooks/xendit", payments.HandleWebhook(repo))
	http.HandleFunc("/api/bookings", handleGetBookings(repo))
	http.HandleFunc("/api/inventory", handleGetInventory(repo))
	
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

func handleGetBookings(repo db.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Enable CORS for the React frontend
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET")
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		bookings, err := repo.GetAllBookings()
		if err != nil {
			log.Printf("Error fetching bookings: %v", err)
			http.Error(w, "Failed to fetch bookings", http.StatusInternalServerError)
			return
		}

		// If no bookings, return empty array instead of null
		if bookings == nil {
			bookings = []db.Booking{}
		}

		if err := json.NewEncoder(w).Encode(bookings); err != nil {
			log.Printf("Error encoding bookings: %v", err)
		}
	}
}

type RoomInventory struct {
	RoomType  string `json:"room_type"`
	Available int    `json:"available"`
	Total     int    `json:"total"`
}

func handleGetInventory(repo db.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Content-Type", "application/json")

		if r.Method == http.MethodOptions {
			w.Header().Set("Access-Control-Allow-Methods", "GET")
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		roomTypes := []string{"standard", "deluxe", "suite"}
		var inventory []RoomInventory

		for _, rt := range roomTypes {
			available, err := cache.GetAvailableRooms(repo, rt)
			if err != nil {
				log.Printf("Error getting availability for %s: %v", rt, err)
				available = 0
			}

			total, err := repo.GetTotalCapacity(rt)
			if err != nil {
				log.Printf("Error getting total capacity for %s: %v", rt, err)
				total = 0
			}

			inventory = append(inventory, RoomInventory{
				RoomType:  rt,
				Available: available,
				Total:     total,
			})
		}

		if err := json.NewEncoder(w).Encode(inventory); err != nil {
			log.Printf("Error encoding inventory: %v", err)
		}
	}
}
