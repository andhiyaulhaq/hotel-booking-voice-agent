# Phase 2: The Brain (Python AI & RAG)

## Objective
Build the stateless LangChain agent that handles conversational reasoning, semantic retrieval (RAG), and executes database operations by acting as a gRPC client to the Go service.

---

## 1. Project Structure
This phase will create the following structure:
```text
brain/
├── data/
│   └── hotel_policies.txt    # Raw text knowledge base
├── src/
│   ├── main.py               # gRPC Server & Entry point
│   ├── agent.py              # LangGraph orchestration & system prompt
│   ├── tools.py              # LangChain tools (gRPC clients to Go)
│   ├── rag.py                # FAISS indexing & retrieval logic
│   └── proto/                # Compiled Python Protobuf files
├── pyproject.toml
└── requirements.txt
```

## 2. Environment Setup
```bash
mkdir -p brain/src/proto brain/data
cd brain
uv init
uv add langchain langchain-openai langgraph faiss-cpu grpcio grpcio-tools python-dotenv
```

## 3. RAG Implementation (FAISS)
1. **Knowledge Base:** Create `data/hotel_policies.txt` containing information on check-in times, pet policies, parking, and amenities.
2. **Vector Store:** Implement `src/rag.py`.
   - Load text documents using `TextLoader`.
   - Chunk text using `RecursiveCharacterTextSplitter`.
   - Generate embeddings via `OpenAIEmbeddings`.
   - Build and save the in-memory `FAISS` index.

## 4. Tool Binding (gRPC Clients)
Implement `src/tools.py`. These tools are bound to the LangChain agent but execute via gRPC.

- `check_availability(room_type: str)`: Opens a gRPC channel to Go, calls `CheckAvailability`.
- `initiate_checkout(guest_name: str, room_type: str, nights: int)`: Calls `InitiateCheckout` on Go, returns the Xendit QRIS URL to the agent (so the agent can tell the user to look at their screen).
- `answer_hotel_question(query: str)`: Queries the local FAISS vector store.

## 5. Agent Orchestration (LangGraph)
Implement `src/agent.py`.

### System Prompt
> "You are a luxury hotel concierge for The Grand AI Hotel. Your goal is to help guests book rooms and answer questions.
> Use `answer_hotel_question` for policy queries. 
> Use `check_availability` before offering a room. 
> When the user confirms, use `initiate_checkout`. Do not ask for credit card numbers."

### State Graph
Use `langgraph` to create a compiled agent with `InMemorySaver` to maintain conversational threads, allowing the agent to remember context (e.g., how many nights the guest requested earlier in the conversation).

## 6. gRPC Server Wrapper
Implement `src/main.py`.
Create a gRPC server that listens for incoming text transcripts from Go. When a transcript arrives, it pushes it into the LangGraph executor and streams the LLM's text response back over the gRPC stream to Go.

---

## 7. Test Scenarios

### Automated Tests (`pytest`)
- **RAG Tests:** Write tests in `src/rag.py` to ensure queries like "pets allowed?" successfully retrieve the relevant chunks from FAISS.
- **Tool Tests:** Mock the gRPC channel in `src/tools.py` and verify that the tools return the expected JSON structures.

### Manual Verification
1. Run `python src/main.py` to start the Python gRPC server.
2. Write a small temporary `client.py` script that sends a mock transcript ("I want to book a suite") to the local Python gRPC server.
3. Verify the LangChain agent intercepts it, attempts to call `check_availability`, and streams a response back.
