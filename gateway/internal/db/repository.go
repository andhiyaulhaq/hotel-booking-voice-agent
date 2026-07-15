package db

import (
	"database/sql"
	"fmt"
)

type BookingRepository interface {
	GetAvailableRooms(roomType string) (int, error)
	CreateBooking(guestName, roomType string, nights int) (int64, error)
	ConfirmBookingStatus(invoiceID string) error
	GetTotalCapacity(roomType string) (int, error)
}

type SQLiteRepository struct {
	db *sql.DB
}

func NewSQLiteRepository(db *sql.DB) *SQLiteRepository {
	return &SQLiteRepository{db: db}
}

// GetTotalCapacity gets the maximum capacity for a specific room type
func (r *SQLiteRepository) GetTotalCapacity(roomType string) (int, error) {
	var total int
	err := r.db.QueryRow("SELECT total_capacity FROM rooms WHERE room_type = ?", roomType).Scan(&total)
	if err != nil {
		if err == sql.ErrNoRows {
			return 0, fmt.Errorf("room type %s not found", roomType)
		}
		return 0, err
	}
	return total, nil
}

// GetAvailableRooms calculates availability by taking total capacity minus confirmed bookings
func (r *SQLiteRepository) GetAvailableRooms(roomType string) (int, error) {
	total, err := r.GetTotalCapacity(roomType)
	if err != nil {
		return 0, err
	}

	var booked int
	// We count both confirmed and pending_payment as occupying a room temporarily to avoid double booking
	err = r.db.QueryRow(`
		SELECT COUNT(*) FROM bookings 
		WHERE room_type = ? AND status IN ('confirmed', 'pending_payment')
	`, roomType).Scan(&booked)
	
	if err != nil && err != sql.ErrNoRows {
		return 0, err
	}

	available := total - booked
	if available < 0 {
		available = 0
	}
	
	return available, nil
}

// CreateBooking inserts a new booking with pending status
func (r *SQLiteRepository) CreateBooking(guestName, roomType string, nights int) (int64, error) {
	// First ensure there is availability
	available, err := r.GetAvailableRooms(roomType)
	if err != nil {
		return 0, err
	}
	
	if available <= 0 {
		return 0, fmt.Errorf("no available rooms of type %s", roomType)
	}

	result, err := r.db.Exec(`
		INSERT INTO bookings (guest_name, room_type, nights, status) 
		VALUES (?, ?, ?, 'pending_payment')
	`, guestName, roomType, nights)
	
	if err != nil {
		return 0, err
	}
	
	return result.LastInsertId()
}

// ConfirmBookingStatus updates the booking status when Xendit confirms payment
func (r *SQLiteRepository) ConfirmBookingStatus(invoiceID string) error {
	result, err := r.db.Exec(`
		UPDATE bookings SET status = 'confirmed' 
		WHERE xendit_invoice_id = ?
	`, invoiceID)
	
	if err != nil {
		return err
	}
	
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return fmt.Errorf("no booking found with invoice ID %s", invoiceID)
	}
	
	return nil
}

// AssociateInvoice associates a generated Xendit invoice ID with a booking
func (r *SQLiteRepository) AssociateInvoice(bookingID int64, invoiceID string) error {
	_, err := r.db.Exec(`
		UPDATE bookings SET xendit_invoice_id = ? 
		WHERE id = ?
	`, invoiceID, bookingID)
	
	return err
}
