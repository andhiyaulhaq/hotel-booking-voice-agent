# Phase 1: The Foundation (Data & State)

## Objective
Establish the state layer and communication contract before introducing AI. The Go microservice (`gateway/`) acts as the single source of truth for database records and in-memory caching.

---

## 1. Project Structure
This phase will create the following structure:
```text
gateway/
├── cmd/
│   └── server/
│       └── main.go           # Entry point (stubbed)
├── internal/
│   ├── db/
│   │   ├── db.go             # SQLite connection & schema init
│   │   └── repository.go     # BookingRepository implementation
│   ├── cache/
│   │   └── redis.go          # Cache-Aside logic
│   └── grpc/
│       └── server.go         # gRPC server implementation (stubbed)
├── proto/
│   ├── service.proto         # Protobuf definitions
│   └── service.pb.go         # Generated Go code (after protoc)
├── go.mod
└── go.sum
```

## 2. Directory Setup & Initialization
```bash
mkdir -p gateway/cmd/server gateway/internal/db gateway/internal/cache gateway/internal/grpc gateway/proto
cd gateway
go mod init github.com/username/hotel-gateway
go get github.com/mattn/go-sqlite3 github.com/redis/go-redis/v9 google.golang.org/grpc google.golang.org/protobuf
```

## 3. Database Schema (SQLite)
Implement the `internal/db` package.

### Tables
1. **rooms**
   - `id` (INTEGER PRIMARY KEY)
   - `room_type` (TEXT UNIQUE) - e.g., 'standard', 'deluxe', 'suite'
   - `total_capacity` (INTEGER)
   - `price_per_night` (INTEGER)
   
2. **bookings**
   - `id` (INTEGER PRIMARY KEY)
   - `guest_name` (TEXT)
   - `room_type` (TEXT)
   - `nights` (INTEGER)
   - `status` (TEXT) - 'pending_payment', 'confirmed', 'cancelled'
   - `xendit_invoice_id` (TEXT)

### Repository Interface
```go
type BookingRepository interface {
    GetAvailableRooms(roomType string) (int, error)
    CreateBooking(guestName, roomType string, nights int) (int64, error)
    ConfirmBookingStatus(invoiceID string) error
}
```

## 4. Cache Management (Redis)
Implement the `internal/cache` package using the **Cache-Aside** pattern.

- **Keys:** `hotel:availability:<room_type>`
- **TTL:** 15 minutes (or flushed upon booking).
- **Flow:** If Python asks for availability, Go checks Redis. If miss, Go queries SQLite, populates Redis, and returns.

## 5. Protobuf Contract (gRPC)
Create `gateway/proto/service.proto` and compile it for both Go and Python.

```protobuf
syntax = "proto3";
package hotelagent;

service HotelStateService {
  rpc CheckAvailability (AvailabilityRequest) returns (AvailabilityResponse);
  rpc InitiateCheckout (CheckoutRequest) returns (CheckoutResponse);
}

message AvailabilityRequest { string room_type = 1; }
message AvailabilityResponse { bool is_available = 1; int32 available_count = 2; }
message CheckoutRequest { string guest_name = 1; string room_type = 2; int32 nights = 3; }
message CheckoutResponse { string invoice_url = 1; string invoice_id = 2; }
```

---

## 6. Test Scenarios

### Automated Tests (`go test`)
- **DB Tests:** Write `internal/db/repository_test.go` to test SQLite inserts and seed data checking.
- **Cache Tests:** Write `internal/cache/redis_test.go` using a mock Redis server (or local docker instance) to verify Cache-Aside logic (cache misses trigger DB reads).

### Manual Verification
1. Run `protoc` to generate Go and Python bindings. Verify no compilation errors.
2. Run the Go server main file and verify it successfully connects to local SQLite and Redis without panicking.
