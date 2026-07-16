import sys
from pathlib import Path
import pytest
import grpc

sys.path.append(str(Path(__file__).parent.parent / "src"))
import tools
import service_pb2

def test_check_availability_success(mocker):
    # Mock the gRPC stub
    mock_response = service_pb2.AvailabilityResponse(
        is_available=True,
        available_count=2
    )
    mock_stub = mocker.Mock()
    mock_stub.CheckAvailability.return_value = mock_response
    
    mocker.patch('tools.stub', mock_stub)
    
    result = tools.check_availability.invoke({"room_type": "suite"})
    assert "Yes, we have 2 suite room(s) available." in result

def test_check_availability_failure(mocker):
    mock_response = service_pb2.AvailabilityResponse(
        is_available=False,
        error_message="Sorry, fully booked."
    )
    mock_stub = mocker.Mock()
    mock_stub.CheckAvailability.return_value = mock_response
    
    mocker.patch('tools.stub', mock_stub)
    
    result = tools.check_availability.invoke({"room_type": "suite"})
    assert "Sorry, fully booked." in result

def test_initiate_checkout_success(mocker):
    mock_response = service_pb2.CheckoutResponse(
        invoice_url="http://xendit.com/test",
        invoice_id="inv_123",
        error_message=""
    )
    mock_stub = mocker.Mock()
    mock_stub.InitiateCheckout.return_value = mock_response
    
    mocker.patch('tools.stub', mock_stub)
    
    result = tools.initiate_checkout.invoke({
        "guest_name": "John Doe", 
        "room_type": "standard", 
        "nights": 3
    })
    
    assert "Successfully reserved" in result
