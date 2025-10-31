package core

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"time"

	"turn-tracker/backend/types"

	"github.com/gorilla/websocket"
)

// MessageHandler is a function that handles incoming messages
type MessageHandler func(hub *Hub, client *Client, msg *types.Message)

const (
	// Time allowed to write a message to the peer.
	writeWait = 10 * time.Second
	// Time allowed to read the next pong message from the peer.
	pongWait = 60 * time.Second
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10
	// Maximum message size allowed from peer.
	maxMessageSize = 64 * 1024 // 64KB (reduced from 512KB to limit memory usage)
)

type Client struct {
	Hub            *Hub
	Conn           *websocket.Conn
	Send           chan []byte
	RoomID         string
	ClientID       string
	DisplayName    string // User's display name
	Color          string // Hex color code (e.g., "#FF5733")
	MessageHandler MessageHandler
}

// GenerateClientID generates a unique client ID
func GenerateClientID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (c *Client) ReadPump() {
	defer func() {
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		// Fast path: Extract message type first for early validation
		// This allows us to quickly identify malformed messages without full JSON parse
		msgType, typeFound := types.FastParseMessageType(messageBytes)
		if !typeFound || msgType == "" {
			// Fast path indicates malformed message, reject early
			errorMsg, _ := types.NewErrorMessage("Invalid message format")
			c.Send <- errorMsg
			continue
		}

		// Parse full message for routing (still needed for Data field)
		var msg types.Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			errorMsg, _ := types.NewErrorMessage("Invalid message format")
			c.Send <- errorMsg
			continue
		}

		// Route message by type
		if c.MessageHandler != nil {
			c.MessageHandler(c.Hub, c, &msg)
		}
	}
}

func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			n := len(c.Send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.Send)
			}

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
