package db

import (
	"testing"
)

func TestSQLiteRepository(t *testing.T) {
	err := InitDB(":memory:")
	if err != nil {
		t.Fatalf("Failed to init in-memory DB: %v", err)
	}
	defer DB.Close()

	repo := NewSQLiteRepository(DB)

	// Test 1: Seed data checking
	capacity, err := repo.GetTotalCapacity("standard")
	if err != nil {
		t.Errorf("Expected to find 'standard' room, got err: %v", err)
	}
	if capacity != 10 {
		t.Errorf("Expected capacity 10, got %d", capacity)
	}

	// Test 2: GetAvailableRooms
	available, err := repo.GetAvailableRooms("standard")
	if err != nil {
		t.Errorf("Failed to get available rooms: %v", err)
	}
	if available != 10 {
		t.Errorf("Expected 10 available, got %d", available)
	}

	// Test 3: CreateBooking
	bookingID, err := repo.CreateBooking("John Doe", "standard", 3)
	if err != nil {
		t.Errorf("Failed to create booking: %v", err)
	}
	if bookingID <= 0 {
		t.Errorf("Expected valid booking ID, got %d", bookingID)
	}

	// Test 4: Check availability after booking (status pending_payment)
	available, err = repo.GetAvailableRooms("standard")
	if err != nil {
		t.Errorf("Failed to get available rooms after booking: %v", err)
	}
	if available != 9 {
		t.Errorf("Expected 9 available, got %d", available)
	}

	// Test 5: Confirm booking
	err = repo.AssociateInvoice(bookingID, "inv_123")
	if err != nil {
		t.Errorf("Failed to associate invoice: %v", err)
	}
	err = repo.ConfirmBookingStatus("inv_123")
	if err != nil {
		t.Errorf("Failed to confirm booking: %v", err)
	}

	// Test 6: Check availability after confirming (should still be 9)
	available, err = repo.GetAvailableRooms("standard")
	if err != nil {
		t.Errorf("Failed to get available rooms after confirmation: %v", err)
	}
	if available != 9 {
		t.Errorf("Expected 9 available, got %d", available)
	}
}
