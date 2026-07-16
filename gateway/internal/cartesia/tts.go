package cartesia

import (
	"encoding/json"
	"fmt"
	"log"

	"github.com/gorilla/websocket"
)

// TTSClient handles the WebSocket connection to Cartesia TTS
type TTSClient struct {
	conn   *websocket.Conn
	Audio  chan []byte
	apiKey string
}

// NewTTSClient initializes a new Cartesia TTS connection
func NewTTSClient(apiKey, voiceID string) (*TTSClient, error) {
	url := fmt.Sprintf("wss://api.cartesia.ai/v1/tts/websocket?api_key=%s", apiKey)
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to Cartesia TTS: %w", err)
	}

	client := &TTSClient{
		conn:   conn,
		Audio:  make(chan []byte),
		apiKey: apiKey,
	}

	// Initialize the TTS context with the voice ID
	initMsg := map[string]interface{}{
		"model_id": "sonic-english",
		"voice": map[string]string{
			"mode": "id",
			"id":   voiceID,
		},
		"output_format": map[string]interface{}{
			"container":   "raw",
			"encoding":    "pcm_s16le",
			"sample_rate": 16000,
		},
	}
	if err := conn.WriteJSON(initMsg); err != nil {
		return nil, fmt.Errorf("failed to init TTS: %w", err)
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
			// Extract base64 audio and send to channel
			if typ, ok := payload["type"].(string); ok && typ == "chunk" {
				if data, ok := payload["data"].(string); ok {
					// We'll pass the base64 string downstream as bytes
					c.Audio <- []byte(data)
				}
			}
		}
	}
}

func (c *TTSClient) SendText(text string) error {
	payload := map[string]interface{}{
		"transcript": text,
		"continue":   true,
	}
	return c.conn.WriteJSON(payload)
}

func (c *TTSClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
