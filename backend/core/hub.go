package core

import (
	"context"
	"log"
	"sync"
	"sync/atomic"
	"time"
)

const (
	// MaxConnections limits the total number of concurrent connections
	MaxConnections = 10000
	// MaxConnectionsPerIP limits connection attempts per IP per window
	MaxConnectionsPerIP = 20 // Limit per IP
)

// DisconnectedClient stores client data when they disconnect
type DisconnectedClient struct {
	ClientID       string
	DisplayName    string
	Color          string
	TotalTurnTime  int64
	LastRoomID     string
	DisconnectedAt time.Time
}

type Hub struct {
	mu    sync.RWMutex // Protects rooms map for reads from cleanup goroutine
	rooms map[string]*Room
	// Registered clients.
	clients map[*Client]bool
	// Register requests from the clients.
	Register chan *Client
	// Unregister requests from clients.
	Unregister chan *Client
	// OnPlayerLeft callback for when a player leaves
	OnPlayerLeft func(roomID, clientID string, message []byte)
	// OnTurnEnded callback for when a turn ends (due to disconnect)
	OnTurnEnded func(roomID string)
	// Current connection count (atomic)
	currentConnections int32
	// Disconnected clients (for reconnection)
	disconnectedClients map[string]*DisconnectedClient
	disconnectedMu      sync.RWMutex // Protects disconnectedClients map
	ipConnections       map[string]int32
	ipMu                sync.RWMutex
	// Shutdown coordination
	shutdownCtx    context.Context
	shutdownCancel context.CancelFunc
	cleanupDone    sync.WaitGroup
}

// NewHub creates a new hub
func NewHub() *Hub {
	ctx, cancel := context.WithCancel(context.Background())
	return &Hub{
		rooms:               make(map[string]*Room),
		clients:             make(map[*Client]bool),
		disconnectedClients: make(map[string]*DisconnectedClient),
		Register:            make(chan *Client, 100), // Buffered to prevent blocking
		Unregister:          make(chan *Client, 100), // Buffered to prevent blocking
		currentConnections:  0,
		ipConnections:       make(map[string]int32),
		shutdownCtx:         ctx,
		shutdownCancel:      cancel,
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
		case <-h.shutdownCtx.Done():
			log.Println("Hub.Run() shutting down...")
			return
		case client := <-h.Register:
			h.handleRegister(client)

		case client := <-h.Unregister:
			h.handleUnregister(client)
		}
	}
}

// Shutdown gracefully shuts down the hub and all its goroutines
func (h *Hub) Shutdown() {
	log.Println("Hub shutdown initiated...")

	// Signal all goroutines to stop
	h.shutdownCancel()

	// Close all client connections
	h.mu.RLock()
	clients := make([]*Client, 0, len(h.clients))
	for client := range h.clients {
		clients = append(clients, client)
	}
	h.mu.RUnlock()

	// Cancel all client contexts and close connections
	for _, client := range clients {
		if client.Cancel != nil {
			client.Cancel()
		}
		if client.Conn != nil {
			client.Conn.Close()
		}
	}

	// Wait for cleanup goroutines to finish
	h.cleanupDone.Wait()

	log.Println("Hub shutdown complete")
}

// GetRoom returns a room if it exists, nil otherwise (thread-safe for reads)
func (h *Hub) GetRoom(roomID string) *Room {
	h.mu.RLock()
	defer h.mu.RUnlock()
	if room, exists := h.rooms[roomID]; exists {
		return room
	}
	return nil
}

// AddRoom adds a room to the hub
// Note: This is called from hub.Run() goroutine, so no lock needed
func (h *Hub) AddRoom(roomID string, room *Room) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if _, exists := h.rooms[roomID]; exists {
		return // Room already exists
	}
	h.rooms[roomID] = room
}

// DeleteRoom removes a room from the hub
// Note: This should only be called from hub.Run() goroutine or with proper locking
func (h *Hub) DeleteRoom(roomID string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	delete(h.rooms, roomID)
}

// RoomExists checks if a room exists (thread-safe for reads)
func (h *Hub) RoomExists(roomID string) bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	_, exists := h.rooms[roomID]
	return exists
}

// HasDisconnectedClients checks if there are any disconnected clients for a room
// O(n) where n is number of disconnected clients (typically small due to 5min TTL)
func (h *Hub) HasDisconnectedClients(roomID string) bool {
	h.disconnectedMu.RLock()
	defer h.disconnectedMu.RUnlock()

	for _, disconnected := range h.disconnectedClients {
		if disconnected.LastRoomID == roomID {
			return true
		}
	}
	return false
}

// TryRegisterIP attempts to register a connection for a specific IP
// Returns false if IP is at limit or global limit reached
func (h *Hub) TryRegisterIP(ip string) bool {
	if ip == "" {
		return false // Reject connections without IP
	}

	// First check global limit
	if !h.TryRegister() {
		return false
	}

	// Check IP limit
	h.ipMu.Lock()
	defer h.ipMu.Unlock()

	current := h.ipConnections[ip]
	if current >= MaxConnectionsPerIP {
		// Rollback global increment
		atomic.AddInt32(&h.currentConnections, -1)
		return false
	}

	h.ipConnections[ip] = current + 1
	return true
}

// UnregisterIP decrements the IP connection count
func (h *Hub) UnregisterIP(ip string) {
	if ip == "" {
		return
	}

	h.ipMu.Lock()
	defer h.ipMu.Unlock()

	if count := h.ipConnections[ip]; count > 0 {
		if count == 1 {
			delete(h.ipConnections, ip)
		} else {
			h.ipConnections[ip] = count - 1
		}
	}
}

// RemoveClientFromRoom removes a client from a room and sends appropriate notifications
// This centralizes the logic for OnTurnEnded and OnPlayerLeft callbacks
// Parameters:
//   - roomID: The room to remove the client from
//   - clientID: The client ID to remove
//   - reason: Optional reason string for logging (e.g., "intentional leave", "disconnect", "moved to another room")
func (h *Hub) RemoveClientFromRoom(roomID, clientID, reason string) {
	h.mu.RLock()
	room := h.rooms[roomID]
	h.mu.RUnlock()

	if room == nil {
		return
	}

	// Remove client from room (handles turn cleanup internally)
	hadCurrentTurn, isEmpty := room.RemoveClient(clientID)

	// If client had current turn, notify that turn ended (even if room becomes empty)
	if hadCurrentTurn && h.OnTurnEnded != nil {
		h.OnTurnEnded(roomID)
	}

	if isEmpty {
		log.Printf("Room %s is now empty (will be cleaned up by scheduled task)", roomID)
		return
	}

	// Room is not empty - notify other players that this player left
	if h.OnPlayerLeft != nil {
		h.OnPlayerLeft(roomID, clientID, nil)
	}

	if reason != "" {
		log.Printf("Client %s removed from room %s: %s", clientID, roomID, reason)
	}
}
