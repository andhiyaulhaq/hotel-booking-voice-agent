package cartesia

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"

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
	params := url.Values{}
	params.Add("api_key", apiKey)
	params.Add("cartesia_version", "2024-03-01")
	params.Add("model", "ink-2")
	params.Add("sample_rate", "16000")
	params.Add("encoding", "pcm_s16le")

	wsUrl := fmt.Sprintf("wss://api.cartesia.ai/stt/turns/websocket?%s", params.Encode())
	conn, _, err := websocket.DefaultDialer.Dial(wsUrl, nil)
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
			if typ, ok := payload["type"].(string); ok && typ == "turn.end" {
				if text, ok := payload["transcript"].(string); ok && text != "" {
					c.Transcript <- text
				}
			}
		}
	}
}

func (c *STTClient) SendAudio(pcmBytes []byte) error {
	// Cartesia STT expects raw binary audio chunks
	return c.conn.WriteMessage(websocket.BinaryMessage, pcmBytes)
}

func (c *STTClient) Close() {
	if c.conn != nil {
		c.conn.Close()
	}
}
