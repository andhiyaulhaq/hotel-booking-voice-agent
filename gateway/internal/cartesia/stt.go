package cartesia

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// STTClient handles the WebSocket connection to Cartesia STT
type STTClient struct {
	conn       *websocket.Conn
	Transcript chan string
	apiKey     string
}

// NewSTTClient initializes a new Cartesia STT connection
func NewSTTClient(apiKey string) (*STTClient, error) {
	url := fmt.Sprintf("wss://api.cartesia.ai/v1/stt?api_key=%s", apiKey)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cartesia STT: %w", err)
	}

	client := &STTClient{
		conn:       conn,
		Transcript: make(chan string),
		apiKey:     apiKey,
	}

	// Start listening for transcripts
	go client.listen()

	return client, nil
}

func (c *STTClient) listen() {
	defer close(c.Transcript)
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Cartesia STT Read Error: %v", err)
			return
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(message, &payload); err == nil {
			// Extract final transcript
			// Note: This payload structure depends on Cartesia's exact API
			if typ, ok := payload["type"].(string); ok && typ == "transcript" {
				if isFinal, ok := payload["is_final"].(bool); ok && isFinal {
					if text, ok := payload["text"].(string); ok {
						c.Transcript <- text
					}
				}
			}
		}
	}
}

func (c *STTClient) SendAudio(pcmBytes []byte) error {
	// Wrap PCM in Cartesia's expected JSON format or send binary if supported
	// Example assuming JSON payload for audio chunks
	payload := map[string]interface{}{
		"audio": pcmBytes, // Usually needs base64 encoding, simplified here
	}
	return c.conn.WriteJSON(payload)
}

func (c *STTClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
