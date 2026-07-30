# Feasibility Study: Migrating to a Custom LangGraph StateGraph

## 1. Executive Summary
Currently, our Hotel Booking Voice Agent uses `langgraph.prebuilt.create_react_agent`. While this is technically LangGraph under the hood, it operates as a generic, linear ReAct (Reasoning + Acting) loop. The AI relies entirely on its system prompt to figure out when to transition between casual conversation, answering FAQs, and executing the booking flow.

This feasibility study evaluates the adoption of a **custom LangGraph `StateGraph`**. A custom graph would allow us to explicitly model the hotel booking flow as a deterministic state machine, dramatically increasing reliability, reducing latency, and preventing the AI from hallucinating tools or jumping out of the booking flow.

**Recommendation:** **Highly Recommended** for production. Moving to a custom StateGraph is highly feasible and will significantly stabilize the voice agent.

---

## 2. Current Architecture vs. Custom LangGraph

### 2.1 Current Architecture (`create_react_agent`)
- **How it works:** The LLM runs in an open-ended loop. At every turn, it decides whether to speak to the user or call a tool (`answer_hotel_question`, `check_availability`, `initiate_checkout`).
- **The Problem:** It relies heavily on prompt engineering (e.g., "NEVER ask for credit card numbers"). If the model ignores the prompt, it might try to check out before checking availability, or hallucinate parameters.

### 2.2 Proposed Architecture (Custom `StateGraph`)
- **How it works:** We define explicit nodes and edges.
  - **Node 1 (`Triage`):** Classifies the user's intent (e.g., "General Question" vs. "Booking").
  - **Node 2 (`QA`):** Dedicated RAG node that enforces concise answers.
  - **Node 3 (`Booking Flow`):** A rigid state machine that collects `room_type`, `nights`, and `name` systematically before allowing checkout.
- **The Solution:** The LLM is constrained. It cannot physically trigger `initiate_checkout` until the graph state confirms all prerequisites are met.

---

## 3. Benefits of a Custom LangGraph

> [!TIP]
> **Deterministic Control**
> Voice AI requires absolute predictability. A custom graph ensures the agent cannot skip steps in the booking process, ensuring the user is always asked for their name and room type.

> [!TIP]
> **Reduced Latency**
> By splitting the agent into smaller nodes (e.g., a dedicated routing node), we can use much smaller, faster LLMs for specific tasks, reducing the "time-to-first-byte" of the voice response.

> [!TIP]
> **Granular Error Handling**
> If a gRPC call to the Go Gateway fails, LangGraph can automatically route to an `ErrorRecovery` node to gracefully apologize to the user and retry, rather than exposing raw errors.

---

## 4. Drawbacks & Risks

> [!WARNING]
> **Increased Complexity**
> A custom StateGraph requires explicitly defining state schemas (`TypedDict`), nodes (Python functions), and conditional edges. The codebase in `brain/src/` will become more verbose.

> [!WARNING]
> **Development Overhead**
> Debugging state transitions in a complex graph can be trickier than debugging a single prompt. We would need to implement LangSmith tracing to monitor the graph's execution path effectively.

---

## 5. Implementation Roadmap

If approved, the migration would follow these steps:

1. **Define the Graph State:**
   ```python
   class AgentState(TypedDict):
       messages: Annotated[list, add_messages]
       intent: str
       booking_details: dict # {room_type, nights, name}
   ```
2. **Build Nodes:**
   - `triage_node`: LLM decides the user's intent.
   - `rag_node`: Executes the vector search and summarizes.
   - `booking_node`: Collects missing slot values.
   - `checkout_node`: Triggers the gRPC tool.
3. **Define Edges:** Wire the nodes together with conditional routing based on the state.
4. **Compile:** `graph = workflow.compile(checkpointer=memory)`
5. **Hot Swap:** Replace `create_react_agent` in `agent.py` with the compiled custom graph.

## 6. Conclusion
Adopting a custom LangGraph `StateGraph` is the definitive next step for making this Hotel Booking Voice Agent production-ready. The transition will take approximately 1-2 days of engineering effort but will yield an infinitely more robust, controllable, and professional AI Concierge.
