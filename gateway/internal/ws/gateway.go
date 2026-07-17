package ws

import (
	"context"
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"fmt"
	"time"

	"github.com/gorilla/websocket"
	"github.com/hotel-voice-agent/gateway/internal/cartesia"
	grpcserver "github.com/hotel-voice-agent/gateway/internal/grpc"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all for demo
	},
}

type ClientMessage struct {
	Type string `json:"type"`
	Data string `json:"data"` // base64 pcm or text
	URL  string `json:"url,omitempty"`
}

// WebSocketHandler struct to inject dependencies
type WebSocketHandler struct {
	AgentClient *grpcserver.AgentClient
}

func NewWebSocketHandler(agentClient *grpcserver.AgentClient) *WebSocketHandler {
	return &WebSocketHandler{AgentClient: agentClient}
}

// HandleConnections upgrades the HTTP connection and routes audio/events
func (wh *WebSocketHandler) HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Error: %v", err)
		return
	}
	defer ws.Close()

	cartesiaKey := os.Getenv("CARTESIA_API_KEY")
	if cartesiaKey == "" {
		log.Println("Warning: CARTESIA_API_KEY not set. Audio streaming will fail.")
	}

	// Initialize Cartesia Clients
	stt, err := cartesia.NewSTTClient(cartesiaKey)
	if err != nil {
		log.Printf("Cartesia STT init error: %v", err)
		return
	}
	defer stt.Close()

	// Default voice ID for the concierge.
	tts, err := cartesia.NewTTSClient(cartesiaKey, "79a125e8-cd45-4c13-8a67-188112f4dd22")
	if err != nil {
		log.Printf("Cartesia TTS init error: %v", err)
		return
	}
	defer tts.Close()

	// Goroutine to send TTS audio back to frontend
	go func() {
		for audioData := range tts.Audio {
			msg := ClientMessage{
				Type: "audio_out",
				Data: string(audioData),
			}
			ws.WriteJSON(msg)
		}
	}()

	// Unique session ID for this WS connection
	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())

	// Goroutine to send STT transcripts to Python Brain
	go func() {
		for transcript := range stt.Transcript {
			log.Printf("Guest: %s", transcript)
			
			var aiResponse string
			
			// Stream text back to Cartesia TTS
			chunkHandler := func(chunk string) {
				aiResponse += chunk
				tts.SendText(chunk)
			}
			
			// Call the Python Brain via gRPC
			err := wh.AgentClient.StreamTranscript(context.Background(), sessionID, transcript, chunkHandler)
			if err != nil {
				log.Printf("Error streaming transcript to python brain: %v", err)
			}
			
			log.Printf("AI Concierge: %s", aiResponse)
		}
	}()

	// Main read loop from frontend
	for {
		var msg ClientMessage
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Client disconnected: %v", err)
			break
		}

		if msg.Type == "audio_in" {
			// Forward PCM bytes to Cartesia STT
			decoded, err := base64.StdEncoding.DecodeString(msg.Data)
			if err != nil {
				log.Printf("Failed to decode audio base64: %v", err)
				continue
			}
			stt.SendAudio(decoded)
		}
	}
}
