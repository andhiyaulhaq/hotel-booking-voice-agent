from typing import Literal
from langchain_openai import ChatOpenAI
from langchain_core.messages import SystemMessage
from langgraph.graph import StateGraph, START, END, MessagesState
from langgraph.prebuilt import ToolNode
from langgraph.checkpoint.memory import InMemorySaver
from tools import TOOLS
import os
from dotenv import load_dotenv

load_dotenv()

SYSTEM_PROMPT = """
You are a luxury hotel concierge for The Grand AI Hotel. 
Your primary goal is to assist guests with answering questions about the hotel and booking rooms.

Guidelines:
1. Always be polite, concise, and professional. You are speaking to the guest over a voice interface, so keep your sentences short and conversational.
2. Use the `answer_hotel_question` tool if the user asks about policies, parking, amenities, or local recommendations.
3. If the user wants to book a room, you MUST use the `check_availability` tool first to ensure the room is available. The available room types are 'standard', 'deluxe', and 'suite'.
4. If they decide to book, ask for their name and how many nights they will stay (if you don't already know).
5. Once you have their name, room type, and nights, use the `initiate_checkout` tool. 
6. NEVER ask for credit card numbers or payment details verbally. Once you trigger `initiate_checkout`, tell the user that a secure payment QR code has appeared on their screen for them to scan.
7. CRITICAL: When using `answer_hotel_question`, you MUST synthesize the answer into 1-2 natural spoken sentences. NEVER output raw markdown (like ## or **) or bulleted lists, as this sounds terrible when spoken by the TTS engine.

Example Booking Flow:
User: "I need a suite for 2 nights."
Agent (Internal): *Calls check_availability("suite")*
Agent: "I see we have suites available. May I have your name to hold the reservation?"
User: "John Doe."
Agent (Internal): *Calls initiate_checkout("John Doe", "suite", 2)*
Agent: "Perfect, John. I've reserved the suite. A secure QRIS code has just appeared on your screen. You can scan it with your GoPay or mobile banking app to complete the payment."
"""

def create_concierge_agent():
    # We use a fast model suitable for voice interactions
    llm = ChatOpenAI(
        model="free-combo",
        temperature=0.3,
        base_url="http://localhost:20128/v1",
        api_key=os.getenv("NINEROUTER_API_KEY", "dummy-key")
    )
    
    # Bind the LLM with our gRPC and RAG tools
    llm_with_tools = llm.bind_tools(TOOLS)
    
    # Define the assistant node
    def assistant(state: MessagesState):
        messages = [SystemMessage(content=SYSTEM_PROMPT)] + state["messages"]
        response = llm_with_tools.invoke(messages)
        return {"messages": [response]}
        
    # Define the routing logic
    def should_continue(state: MessagesState) -> Literal["tools", "__end__"]:
        messages = state["messages"]
        last_message = messages[-1]
        
        # If the LLM makes a tool call, route to the tools node
        if last_message.tool_calls:
            return "tools"
            
        # Otherwise, we are done
        return END

    # Create the graph
    workflow = StateGraph(MessagesState)
    
    # Add nodes
    workflow.add_node("assistant", assistant)
    workflow.add_node("tools", ToolNode(TOOLS))
    
    # Define edges
    workflow.add_edge(START, "assistant")
    workflow.add_conditional_edges("assistant", should_continue)
    workflow.add_edge("tools", "assistant")
    
    # We use InMemorySaver to remember the conversation context per session/thread
    memory = InMemorySaver()
    
    # Compile the graph
    agent_executor = workflow.compile(checkpointer=memory)
    
    return agent_executor
