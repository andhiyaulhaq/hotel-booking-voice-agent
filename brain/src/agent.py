from langchain_openai import ChatOpenAI
from langgraph.prebuilt import create_react_agent
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
        model="gpt-4o-mini",
        temperature=0.3
    )
    
    # Bind the LLM with our gRPC and RAG tools
    # We use InMemorySaver to remember the conversation context per session/thread
    memory = InMemorySaver()
    
    agent_executor = create_react_agent(
        model=llm,
        tools=TOOLS,
        checkpointer=memory,
        state_modifier=SYSTEM_PROMPT
    )
    
    return agent_executor
