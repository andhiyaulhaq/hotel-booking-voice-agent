import grpc
from langchain_core.tools import tool
import sys
from pathlib import Path

# Add proto generated files to path
proto_dir = str(Path(__file__).parent / "proto")
sys.path.append(proto_dir)

import service_pb2
import service_pb2_grpc
from rag import search_hotel_policies

# Configure the gRPC channel to talk to the Go Gateway
# We keep a single channel open for the lifetime of the process
channel = grpc.insecure_channel('localhost:50051')
stub = service_pb2_grpc.HotelStateServiceStub(channel)

@tool
def answer_hotel_question(query: str) -> str:
    """
    Use this tool to answer general questions about the hotel, such as check-in times, 
    pet policies, parking, amenities, and local recommendations.
    """
    return search_hotel_policies(query)

@tool
def check_availability(room_type: str) -> str:
    """
    Check if a specific room type is available for booking. 
    Valid room types are 'standard', 'deluxe', and 'suite'.
    You MUST use this tool before offering a room to a guest.
    """
    try:
        request = service_pb2.AvailabilityRequest(room_type=room_type.lower())
        response = stub.CheckAvailability(request)
        
        if response.is_available:
            return f"Yes, we have {response.available_count} {room_type} room(s) available."
        else:
            msg = response.error_message if response.error_message else f"Sorry, there are no {room_type} rooms available right now."
            return msg
    except grpc.RpcError as e:
        return f"System error checking availability: {e.details()}"

@tool
def initiate_checkout(guest_name: str, room_type: str, nights: int) -> str:
    """
    Finalize a booking and trigger the payment process.
    Use this ONLY when the user explicitly agrees to book the room.
    Do NOT ask for credit card details. This tool generates a secure payment link on their screen.
    """
    try:
        request = service_pb2.CheckoutRequest(
            guest_name=guest_name,
            room_type=room_type.lower(),
            nights=nights
        )
        response = stub.InitiateCheckout(request)
        
        if response.error_message:
            return f"Failed to initiate checkout: {response.error_message}"
            
        # Return a message that the agent can paraphrase to the user
        return (
            f"Successfully reserved the {room_type} room for {guest_name} for {nights} nights. "
            f"A secure payment QRIS code has been generated on their screen."
        )
    except grpc.RpcError as e:
        return f"System error initiating checkout: {e.details()}"

# Export the list of tools for the agent to bind
TOOLS = [answer_hotel_question, check_availability, initiate_checkout]
