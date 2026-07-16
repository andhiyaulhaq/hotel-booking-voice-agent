package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/hotel-voice-agent/gateway/internal/db"
	"github.com/xendit/xendit-go/v6"
	"github.com/xendit/xendit-go/v6/invoice"
)

var xenditClient *xendit.APIClient

func InitXendit() {
	apiKey := os.Getenv("XENDIT_API_KEY")
	if apiKey == "" {
		log.Println("Warning: XENDIT_API_KEY not set. Payments will mock success.")
		return
	}
	xenditClient = xendit.NewClient(apiKey)
	log.Println("Xendit API Client initialized.")
}

// GenerateInvoice creates a new Xendit invoice for a booking
func GenerateInvoice(bookingID int64, guestName string, amount float64) (string, string, error) {
	if xenditClient == nil {
		// Mock Mode
		return fmt.Sprintf("inv_mock_%d", bookingID), "https://checkout.xendit.co/web/mock", nil
	}

	extID := fmt.Sprintf("hotel_booking_%d", bookingID)
	req := xenditClient.InvoiceApi.CreateInvoice(context.Background()).
		CreateInvoiceRequest(invoice.CreateInvoiceRequest{
			ExternalId:  extID,
			Amount:      amount,
		})

	resp, _, err := req.Execute()
	if err != nil {
		return "", "", fmt.Errorf("xendit API error: %w", err)
	}

	return *resp.Id, resp.InvoiceUrl, nil
}

// HandleWebhook processes incoming Xendit webhooks
func HandleWebhook(repo db.BookingRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var payload map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
			http.Error(w, "Bad request", http.StatusBadRequest)
			return
		}

		status, ok := payload["status"].(string)
		invID, idOk := payload["id"].(string)

		if ok && idOk && status == "PAID" {
			// Payment successful, confirm booking in SQLite
			if err := repo.ConfirmBookingStatus(invID); err != nil {
				log.Printf("Failed to confirm booking for invoice %s: %v", invID, err)
			} else {
				log.Printf("Successfully confirmed booking for invoice %s", invID)
				// Note: in a real app, you would also clear the specific redis cache key here
				// but since we only have the roomType in the cache key, we need to fetch the roomType from the booking first.
			}
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}
}
