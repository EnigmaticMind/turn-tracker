package core

import (
	"log"
	"sync/atomic"
)

const (
	// MaxConnections limits the total number of concurrent connections
	MaxConnections = 10000
)

type Hub struct {
	rooms map[string]*Room
	// Registered clients.
	clients map[*Client]bool
	// Inbound messages from the clients.
	broadcast chan []byte
	// Register requests from the clients.
	Register chan *Client
	// Unregister requests from clients.
	Unregister chan *Client
	// OnPlayerLeft callback for when a player leaves
	OnPlayerLeft func(roomID, clientID string, message []byte)
	// Current connection count (atomic)
	currentConnections int32
}

// NewHub creates a new hub
func NewHub() *Hub {
	return &Hub{
		rooms:              make(map[string]*Room),
		clients:            make(map[*Client]bool),
		broadcast:          make(chan []byte),
		Register:           make(chan *Client, 100), // Buffered to prevent blocking
		Unregister:         make(chan *Client, 100), // Buffered to prevent blocking
		currentConnections: 0,
	}
}

// TryRegister attempts to register a new connection, returns false if at limit
// Uses atomic operations to prevent race conditions
func (h *Hub) TryRegister() bool {
	for {
		current := atomic.LoadInt32(&h.currentConnections)
		if current >= MaxConnections {
			return false
		}
		// Try to increment using CompareAndSwap for atomicity
		if atomic.CompareAndSwapInt32(&h.currentConnections, current, current+1) {
			return true
		}
		// Retry if CAS failed (another goroutine modified it)
	}
}

// UnregisterConnection decrements the connection counter
func (h *Hub) UnregisterConnection() {
	atomic.AddInt32(&h.currentConnections, -1)
}

func (h *Hub) Run() {
	for {
		select {
		case client := <-h.Register:
			// Generate client ID if not set
			if client.ClientID == "" {
				client.ClientID = GenerateClientID()
			}
			h.clients[client] = true
			log.Printf("Client registered: %s (ID: %s)", client.Conn.RemoteAddr(), client.ClientID)
			// Room assignment happens via create_room or join_room messages

		case client := <-h.Unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
				h.UnregisterConnection() // Decrement connection counter

				// Remove client from room and notify others
				if room, exists := h.rooms[client.RoomID]; exists {
					delete(room.Clients, client)

					// If room is empty, remove it
					if len(room.Clients) == 0 {
						delete(h.rooms, client.RoomID)
						log.Printf("Room %s deleted (empty)", client.RoomID)
					} else {
						// Call callback if set (for player left message)
						if h.OnPlayerLeft != nil {
							h.OnPlayerLeft(client.RoomID, client.ClientID, nil)
						}
					}
				}

				log.Printf("Client unregistered: %s (ID: %s)", client.Conn.RemoteAddr(), client.ClientID)
			}

		case message := <-h.broadcast:
			// Global broadcast (not currently used, but kept for compatibility)
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// GetRoom returns a room if it exists, nil otherwise
func (h *Hub) GetRoom(roomID string) *Room {
	if room, exists := h.rooms[roomID]; exists {
		return room
	}
	return nil
}

// AddRoom adds a room to the hub
func (h *Hub) AddRoom(roomID string, room *Room) {
	h.rooms[roomID] = room
}

// RoomExists checks if a room exists
func (h *Hub) RoomExists(roomID string) bool {
	_, exists := h.rooms[roomID]
	return exists
}

// BroadcastToRoomExcept broadcasts a message to all clients in a room except the specified client
func (h *Hub) BroadcastToRoomExcept(roomID string, except *Client, message []byte) {
	if room, exists := h.rooms[roomID]; exists {
		for client := range room.Clients {
			if client != except {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(room.Clients, client)
				}
			}
		}
	}
}

// BroadcastToRoom broadcasts a message to all clients in a room
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	if room, exists := h.rooms[roomID]; exists {
		for client := range room.Clients {
			select {
			case client.Send <- message:
			default:
				close(client.Send)
				delete(room.Clients, client)
			}
		}
	}
}
