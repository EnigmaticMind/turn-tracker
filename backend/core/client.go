package core

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"log"
	"sync"
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
	pongWait = 90 * time.Second // Increased from 60 to 90 seconds
	// Send pings to peer with this period. Must be less than pongWait.
	pingPeriod = (pongWait * 9) / 10 // Now ~81 seconds instead of 54
	// Maximum message size allowed from peer.
	maxMessageSize = 64 * 1024 // 64KB
)

type Client struct {
	Hub            *Hub
	Conn           *websocket.Conn
	Ctx            context.Context
	Cancel         context.CancelFunc
	Send           chan []byte
	RoomID         string
	ClientID       string
	DisplayName    string // User's display name
	Color          string // Hex color code (e.g., "#FF5733")
	TotalTurnTime  int64  // Total time spent in turns (in milliseconds)
	MessageHandler MessageHandler
	rateLimit      *clientRateLimit
	rateLimitOnce  sync.Once
	IP             string // Client's IP address (for connection limiting)
}

// messageBufferPool is a pool of reusable byte slices for message buffers
// This reduces allocations and GC pressure
var messageBufferPool = sync.Pool{
	New: func() interface{} {
		// Create new buffer with 1KB initial capacity
		// This is a good default size for most JSON messages
		return make([]byte, 0, 1024)
	},
}

// GetMessageBuffer retrieves a buffer from the pool or creates a new one
// Exported for use in other packages that need to pool message buffers
func GetMessageBuffer() []byte {
	buf := messageBufferPool.Get().([]byte)
	// Reset length but keep capacity (reuse the allocated memory)
	return buf[:0]
}

// PutMessageBuffer returns a buffer to the pool for reuse
// Only pool buffers that are reasonably sized to avoid memory bloat
// Exported for use in other packages
func PutMessageBuffer(buf []byte) {
	// Only pool buffers up to 64KB to avoid holding onto huge buffers
	if cap(buf) <= 64*1024 {
		messageBufferPool.Put(buf)
	}
	// If larger than 64KB, let GC handle it (don't pool)
}

// CopyToPooledBuffer copies data into a pooled buffer
// This is useful after json.Marshal which allocates its own buffer
// Returns a pooled buffer containing the data
func CopyToPooledBuffer(data []byte) []byte {
	buf := GetMessageBuffer()
	// Ensure buffer has enough capacity
	if cap(buf) < len(data) {
		// If pooled buffer is too small, return it to pool and allocate a new one
		PutMessageBuffer(buf)

		// Allocate new buffer with right size (but don't pool oversized buffers)
		if len(data) <= 64*1024 {
			buf = make([]byte, 0, len(data)+1024) // Add some headroom for future messages
		} else {
			// For large messages, just allocate normally (don't pool)
			result := make([]byte, len(data))
			copy(result, data)
			return result
		}
	}
	// Copy data into pooled buffer (reuse capacity, reset length)
	buf = append(buf[:0], data...)
	return buf
}

// GenerateClientID generates a unique client ID
func GenerateClientID() string {
	b := make([]byte, 8)
	rand.Read(b)
	return hex.EncodeToString(b)
}

// SafeSend safely sends a message to the client's Send channel
// Returns true if sent successfully, false if channel is closed or full
func (c *Client) SafeSend(message []byte) bool {
	defer func() {
		if r := recover(); r != nil {
			// Channel closed - expected during client disconnect
			// Return false to indicate send failed
		}
	}()

	select {
	case c.Send <- message:
		return true
	default:
		// Channel full (non-blocking send failed)
		// Note: If channel is closed, the select will panic before checking cases
		// The recover() above handles that case
		return false
	}
}

func (c *Client) ReadPump() {
	defer func() {
		if c.Cancel != nil {
			c.Cancel()
		}
		c.Hub.Unregister <- c
		c.Conn.Close()
	}()

	if !c.CheckRateLimit() {
		errorMsg, _ := types.NewErrorMessage("Rate limit exceeded")
		c.SafeSend(errorMsg)
		return
	}

	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		// Check context before blocking read
		select {
		case <-c.Ctx.Done():
			return
		default:
		}

		_, messageBytes, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error: %v", err)
			}
			break
		}

		var msg types.Message
		if err := json.Unmarshal(messageBytes, &msg); err != nil {
			log.Printf("Error parsing message: %v", err)
			errorMsg, _ := types.NewErrorMessage("Invalid message format")
			c.SafeSend(errorMsg)
			continue
		}

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
		// Note: We don't call cancel here - ReadPump handles that
	}()

	for {
		select {
		case <-c.Ctx.Done():
			// Context cancelled, exit gracefully
			return

		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				// Pool the message before returning
				PutMessageBuffer(message)
				return
			}
			// Write first message
			w.Write(message)

			// Batch loop: grab more messages if available
			// Track all messages for pooling after sending
			const maxBatch = 10
			messages := [maxBatch][]byte{message} // Track all messages in batch
			messageCount := 1

			for messageCount < maxBatch {
				select {
				case <-c.Ctx.Done():
					// Context cancelled during batching
					w.Close()
					// Pool all messages we collected
					for i := 0; i < messageCount; i++ {
						PutMessageBuffer(messages[i])
					}
					return
				case msg, ok := <-c.Send:
					if !ok {
						goto closeWriter
					}
					w.Write([]byte{'\n'})
					w.Write(msg)
					messages[messageCount] = msg
					messageCount++
				default:
					goto closeWriter
				}
			}
		closeWriter:
			if err := w.Close(); err != nil {
				// Pool all messages before returning on error
				for i := 0; i < messageCount; i++ {
					PutMessageBuffer(messages[i])
				}
				return
			}

			// Pool all messages after successful write
			// Messages have been written to the network, safe to reuse buffers
			for i := 0; i < messageCount; i++ {
				PutMessageBuffer(messages[i])
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// When client disconnects or needs to be stopped:
func cleanupClient(client *Client) {
	client.Cancel() // Signals ReadPump and WritePump to stop
	client.Conn.Close()
}
