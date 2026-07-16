package ws

import (
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/hotel-voice-agent/gateway/internal/cartesia"
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

// HandleConnections upgrades the HTTP connection and routes audio/events
func HandleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket Upgrade Error: %v", err)
		return
	}
	defer ws.Close()

	// Initialize Cartesia Clients (assuming keys are set)
	stt, err := cartesia.NewSTTClient("mock-key")
	if err != nil {
		log.Printf("Cartesia STT init error: %v", err)
		return
	}
	defer stt.Close()

	tts, err := cartesia.NewTTSClient("mock-key", "voice-id-here")
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

	// Goroutine to send STT transcripts to Python Brain
	go func() {
		for transcript := range stt.Transcript {
			log.Printf("Final Transcript: %s", transcript)
			// TODO: Send to Python via gRPC
			
			// Mock: Send text directly to TTS for echo
			tts.SendText(transcript)
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
			stt.SendAudio([]byte(msg.Data))
		}
	}
}
