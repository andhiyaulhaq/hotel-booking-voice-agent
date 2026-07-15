# Implementation Plan: Hotel Booking Voice Agent

This document outlines the step-by-step execution to build the hybrid Go/Python Voice AI Hotel Concierge.

---

## 1. Directory Structure

The project is initialized with the following structure:

```text
hotel-booking-voice-agent/
├── gateway/             # Go Microservice (Edge, DB, Cache, Xendit)
│   ├── cmd/             # Main application entrypoint
│   ├── internal/        # Private application code (db, redis, grpc, ws)
│   ├── proto/           # gRPC Protobuf definitions
│   ├── go.mod
│   └── Makefile
├── brain/               # Python Microservice (LangChain, FAISS, Prompts)
│   ├── src/             # Main application code (agent, tools, rag)
│   ├── data/            # Knowledge base documents for FAISS
│   ├── pyproject.toml
│   └── requirements.txt
├── ui/                  # Web Frontend (Guest UI & Admin Dashboard)
│   ├── guest/           # The Booking Portal & Voice overlay
│   └── admin/           # The Manager Dashboard
├── docs/                # Project Documentation
└── docker-compose.yml   # For orchestrating Redis and the services
```

---

## 2. Execution Phases

### Phase 1: The Foundation (Data & State)
Before any AI is involved, we need a rock-solid data layer.
1. **Initialize Go Module:** Setup `gateway/` directory and go module.
2. **SQLite Database:** Implement the `BookingRepository` interface using `mattn/go-sqlite3`. Create tables for `rooms` and `bookings`.
3. **Redis Cache:** Implement Cache-Aside logic using `go-redis` to instantly check room availability.
4. **Protobuf Definitions:** Write `service.proto` defining the gRPC contract between Go and Python (`CheckAvailability`, `InitiateCheckout`, etc.) and compile it for both Go and Python.

### Phase 2: The Brain (Python AI & RAG)
Building the stateless reasoning engine.
1. **Initialize Python Env:** Setup `brain/` with LangChain, FAISS, and gRPC dependencies using `uv` or `pip`.
2. **Knowledge Base:** Create a sample `hotel_policies.txt` in the `data/` folder and generate the local FAISS index.
3. **Tool Creation:** Write LangChain tools that act as gRPC clients (calling Go to check DB state) and RAG tools (querying FAISS).
4. **Agent Orchestration:** Build the LangGraph/LangChain agent with the luxury hotel concierge system prompt.
5. **gRPC Server:** Wrap the Python agent in a gRPC service so Go can send it transcripts.

### Phase 3: The Edge (Go WebSockets & Cartesia)
Connecting the real-time audio and payment flows.
1. **Cartesia Integration:** Build Go WebSocket clients to stream audio to Cartesia STT and receive from Cartesia TTS.
2. **Client Gateway:** Build the main WebSocket server that the Web UI connects to.
3. **Xendit Integration:** Implement Xendit API calls to generate QRIS codes/Virtual Accounts, and create a Webhook endpoint to listen for payment success.
4. **Pipeline Assembly:** Wire the full flow: `Client Audio -> Go -> Cartesia STT -> Python Agent (gRPC) -> Go -> Cartesia TTS -> Client Audio`.

### Phase 4: The Interfaces (Web UI)
Building the visual components.
1. **Guest UI:** Build the booking portal with a glassmorphism "Voice Concierge" overlay. It will connect via WebSocket to stream audio and receive UI-filtering events (like showing a QRIS code).
2. **Admin UI:** Build a simple dashboard connecting via WebSocket to show real-time room availability and confirmed bookings.

---

## 3. Verification Plan

### Automated/Unit Testing
- **Go:** Write table-driven tests for SQLite queries and Redis cache invalidation.
- **Python:** Write unit tests mocking gRPC responses to ensure the LangChain agent calls tools correctly.

### Manual End-to-End Testing
1. Spin up the cluster using `docker-compose up` (Redis) and running the Go and Python servers.
2. Open the Guest UI, click the microphone, and say: *"Do you have any suites for this weekend? Also, what is your pet policy?"*
3. Verify RAG FAISS correctly answers the pet policy.
4. Verify Go Redis correctly checks suite availability.
5. Say: *"Book the suite."*
6. Verify the UI displays the Xendit QRIS code.
7. Manually hit the Go Webhook endpoint simulating a successful payment.
8. Verify the AI verbally confirms the booking and the Admin UI ledger updates instantly.
