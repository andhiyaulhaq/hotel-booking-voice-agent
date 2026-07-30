# Hotel Booking Voice Agent 🏨🎙️

An intelligent, production-grade Voice AI Concierge designed to handle real-time hotel bookings, answer FAQs, and process payments. Built using a high-performance hybrid microservices architecture combining Go and Python.

---

## 🏗️ Architecture

The system is split into two primary microservices that communicate synchronously via **gRPC**:

### 1. Gateway Edge Service (`/gateway`)
Built in **Go (Golang)**, this service acts as the high-throughput edge node facing the client. 
- **WebSocket Streaming:** Handles real-time audio streams from users.
- **REST API:** Serves data endpoints (e.g., `/api/bookings`) for the React Admin Dashboard.
- **State Management:** Manages all persistent state via **SQLite** and implements a Cache-Aside pattern using **Redis** for lightning-fast room availability checks.
- **Payments:** Integrates with the **Xendit** API for processing webhooks and managing booking payments.
- **Performance:** Designed to be extremely low-latency, keeping the voice interactions feeling natural and responsive.

### 2. AI Brain (`/brain`)
Built in **Python**, this stateless service powers the conversational intelligence and decision making.
- **Agentic Reasoning:** Uses a **Custom LangGraph StateGraph** to process incoming transcripts, determine user intents deterministically, and execute function-calling tools.
- **Retrieval-Augmented Generation (RAG):** Integrates **FAISS** to rapidly retrieve hotel policies, local tourism knowledge, and other unstructured data to answer guest questions contextually.
- **Streaming Responses:** Streams AI responses chunk-by-chunk back to the Gateway via gRPC for immediate TTS (Text-to-Speech) processing.

---

## 🚀 Features
- **Real-Time Voice Interaction:** Book rooms entirely through natural voice conversations.
- **Live Inventory Checking:** The AI Brain queries the Go Gateway in real-time to check Redis caches for room availability before confirming bookings.
- **Automated Payments:** Automatically generates payment invoices and confirms bookings asynchronously when payment is received via Xendit webhooks.
- **RAG Powered Knowledge:** Ask the concierge about check-in policies, nearby attractions, or breakfast hours, and it will respond accurately based on vectorized documents.

---

## 🛠️ Tech Stack
- **Languages:** Go 1.25, Python 3.12+, TypeScript
- **Frontend:** React, Vite
- **AI/ML:** Custom LangGraph StateGraph, LangChain, FAISS
- **Database:** SQLite
- **Caching:** Redis (via Docker Compose)
- **RPC Framework:** gRPC & Protocol Buffers
- **Payments:** Xendit Go API

---

## 📦 Getting Started

### Prerequisites
- Go 1.25+
- Python 3.12+ (or `uv` package manager)
- Docker & Docker Compose
- Protoc (Protocol Buffers Compiler)

### 0. Start the Database & Cache (Docker Compose)
The system uses PostgreSQL for persistence and Redis for ultra-low latency state caching. A `docker-compose.yml` file is provided at the root of the project.

Before running Docker Compose, you can define your secure credentials by creating a `.env` file in the root directory:
```env
POSTGRES_USER=my_secure_user
POSTGRES_PASSWORD=my_super_secret_password
POSTGRES_DB=hotel_production_db
```
Then, spin up the infrastructure:
```bash
docker compose up -d
```

> [!WARNING]
> **Troubleshooting Database Authentication**
> If you start the Postgres container *before* creating the `.env` file, Docker will initialize the database volume with the default credentials (`hotel_user`). If you later add the `.env` file and try to connect, you will get a `password authentication failed` error. 
> To fix this, you must wipe the old volume: `docker compose down -v` and then run `docker compose up -d` again.

### 1. Start the Go Gateway
```bash
cd gateway
# Ensure dependencies are tidy
go mod tidy

# Start the edge server (Port 8080 for HTTP/WS, 50051 for gRPC)
go run ./cmd/server
```

### 2. Start the Python AI Brain
```bash
cd brain
# Set up virtual environment and install dependencies
uv venv
source .venv/bin/activate
uv pip install -r pyproject.toml # or equivalent

# Start the gRPC Brain server (Port 50052)
python src/main.py
```

### 3. Start the Web UI (Frontend)
```bash
cd web
# Install dependencies using pnpm
pnpm install

# Start the Vite development server (Port 5173)
pnpm run dev
```

### 4. Running Tests
The Gateway service includes comprehensive tests for database logic and cache management.
```bash
cd gateway
go test ./...
```

---

## 🗺️ Roadmap
- [x] **Phase 1:** Core Data State (SQLite/Redis) & gRPC Contract
- [x] **Phase 2:** Custom LangGraph AI Integration & Voice Endpoints (Cartesia)
- [x] **Phase 3:** Multimodal UI Booking Portal & Admin Dashboard
- [ ] **Phase 4:** Xendit Payment Integration (Sandbox)
- [ ] **Phase 5:** Supabase Integration for scalable cloud database & Auth

---

## 📝 License
This project is an open-source prototype licensed under the MIT License.
