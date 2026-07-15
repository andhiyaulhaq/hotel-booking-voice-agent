package db

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/mattn/go-sqlite3"
)

var DB *sql.DB

func InitDB(dbPath string) error {
	var err error
	DB, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		return fmt.Errorf("failed to open database: %w", err)
	}

	if err := DB.Ping(); err != nil {
		return fmt.Errorf("failed to ping database: %w", err)
	}

	if err := createSchema(); err != nil {
		return err
	}

	if err := seedData(); err != nil {
		return err
	}

	log.Println("Database initialized successfully.")
	return nil
}

func createSchema() error {
	schema := `
	CREATE TABLE IF NOT EXISTS rooms (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		room_type TEXT UNIQUE NOT NULL,
		total_capacity INTEGER NOT NULL,
		price_per_night INTEGER NOT NULL
	);

	CREATE TABLE IF NOT EXISTS bookings (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		guest_name TEXT NOT NULL,
		room_type TEXT NOT NULL,
		nights INTEGER NOT NULL,
		status TEXT NOT NULL,
		xendit_invoice_id TEXT
	);
	`
	_, err := DB.Exec(schema)
	if err != nil {
		return fmt.Errorf("failed to create schema: %w", err)
	}
	return nil
}

func seedData() error {
	var count int
	err := DB.QueryRow("SELECT COUNT(*) FROM rooms").Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check room count: %w", err)
	}

	if count == 0 {
		log.Println("Seeding initial room data...")
		seedQuery := `
		INSERT INTO rooms (room_type, total_capacity, price_per_night) VALUES 
		('standard', 10, 100),
		('deluxe', 5, 200),
		('suite', 2, 500);
		`
		_, err := DB.Exec(seedQuery)
		if err != nil {
			return fmt.Errorf("failed to seed rooms: %w", err)
		}
	}
	return nil
}
