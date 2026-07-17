package cartesia

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

	"github.com/gorilla/websocket"
)

// TTSClient handles the WebSocket connection to Cartesia TTS
type TTSClient struct {
	conn    *websocket.Conn
	Audio   chan []byte
	apiKey  string
	voiceID string
}

// NewTTSClient initializes a new Cartesia TTS connection
func NewTTSClient(apiKey, voiceID string) (*TTSClient, error) {
	params := url.Values{}
	params.Add("api_key", apiKey)
	params.Add("cartesia_version", "2024-03-01") // Or 2025-04-16

	wsUrl := fmt.Sprintf("wss://api.cartesia.ai/tts/websocket?%s", params.Encode())
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cartesia TTS: %w", err)
	}

	client := &TTSClient{
		conn:    conn,
		Audio:   make(chan []byte),
		apiKey:  apiKey,
		voiceID: voiceID,
	}

	go client.listen()

	return client, nil
}

func (c *TTSClient) listen() {
	defer close(c.Audio)
	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			log.Printf("Cartesia TTS Read Error: %v", err)
			return
		}

		var payload map[string]interface{}
		if err := json.Unmarshal(message, &payload); err == nil {
			// Cartesia sends chunk data in the "data" field (base64)
			if data, ok := payload["data"].(string); ok && data != "" {
				c.Audio <- []byte(data)
			}
			if errStr, ok := payload["error"].(string); ok && errStr != "" {
				log.Printf("Cartesia TTS Error from server: %s", errStr)
			}
		}
	}
}

func (c *TTSClient) SendText(text string) error {
	if text == "" {
		return nil
	}

	// Use the same context ID for the duration of this client connection to prevent concurrency limits
	contextID := "ctx_hotel_concierge_" + c.voiceID

	payload := map[string]interface{}{
		"model_id":   "sonic-3",
		"transcript": text,
		"voice": map[string]string{
			"mode": "id",
			"id":   c.voiceID,
		},
		"output_format": map[string]interface{}{
			"container":   "raw",
			"encoding":    "pcm_s16le",
			"sample_rate": 16000,
		},
		"context_id": contextID,
		"continue":   true,
	}
	return c.conn.WriteJSON(payload)
}

func (c *TTSClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
