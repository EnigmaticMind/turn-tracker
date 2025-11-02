package core

import (
	"sync"
	"time"
)

type PeerInfo struct {
	ClientID      string `json:"client_id"`
	DisplayName   string `json:"display_name"`
	Color         string `json:"color"`
	TotalTurnTime int64  `json:"total_turn_time"` // Total time spent in turns (in milliseconds)
}

type Room struct {
	mu            sync.RWMutex // Protects all Room state
	ID            string
	Clients       map[string]*Client
	CreatedBy     string    // clientID of the room creator
	CreatedAt     time.Time // When the room was created (for cleanup)
	CurrentTurn   string    // clientID of the player whose turn it is (empty if no turn active)
	TurnStartTime *int64    // Unix timestamp in nanoseconds when current turn started (nil if no turn active)
	turnSequence  uint64    // Sequence number for turn_changed messages (incremented on each turn change)
}

// NewRoom creates a new room
func NewRoom(id string) *Room {
	return &Room{
		ID:        id,
		Clients:   make(map[string]*Client),
		CreatedAt: time.Now(),
	}
}

// AddClient adds a client to the room (thread-safe)
// Returns true if client was added, false if already exists
func (r *Room) AddClient(client *Client) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.Clients[client.ClientID] != nil {
		return false // Already in room
	}

	r.Clients[client.ClientID] = client
	return true
}

// ListPeerIDs returns all client IDs in the room (kept for backward compatibility)
func (r *Room) ListPeerIDs() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Consider returning nil for empty rooms
	if len(r.Clients) == 0 {
		return nil
	}

	peers := make([]string, 0, len(r.Clients))
	for clientID := range r.Clients {
		peers = append(peers, clientID)
	}
	return peers
}

// ListPeerInfo returns all peer information in the room
func (r *Room) ListPeerInfo() []PeerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	peers := make([]PeerInfo, 0, len(r.Clients))
	for _, client := range r.Clients {
		peerInfo := PeerInfo{
			ClientID:      client.ClientID,
			DisplayName:   client.DisplayName,
			Color:         client.Color,
			TotalTurnTime: client.TotalTurnTime,
		}

		peers = append(peers, peerInfo)
	}

	return peers
}

// GetCurrentTurnInfo returns the peer info for the current turn, or nil if no turn active
func (r *Room) GetCurrentTurnInfo() PeerInfo {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if r.CurrentTurn == "" {
		return PeerInfo{}
	}

	client := r.Clients[r.CurrentTurn]
	if client == nil {
		// Current turn player not found (shouldn't happen, but safe)
		// This can happen if client was removed after turn was set
		return PeerInfo{}
	}

	return PeerInfo{
		ClientID:      client.ClientID,
		DisplayName:   client.DisplayName,
		Color:         client.Color,
		TotalTurnTime: client.TotalTurnTime,
	}
}

// GetCurrentTurn returns the current turn client ID (thread-safe read)
func (r *Room) GetCurrentTurn() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.CurrentTurn
}

// GetTurnStartTime returns the turn start time in milliseconds (thread-safe read)
func (r *Room) GetTurnStartTime() int64 {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if r.TurnStartTime == nil {
		return 0
	}
	// Convert from nanoseconds to milliseconds
	return (*r.TurnStartTime) / int64(time.Millisecond)
}

// GetTurnSequence returns the current turn sequence number and increments it (thread-safe)
// This ensures each turn_changed message has a unique, incrementing sequence number
func (r *Room) GetTurnSequence() uint64 {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.turnSequence++
	return r.turnSequence
}

// SetCurrentTurn sets the current turn to the specified client ID atomically
// Validates expectedCurrentTurn matches before setting (optimistic concurrency)
// Returns true if turn was set, false if validation failed or client not found
func (r *Room) SetCurrentTurn(expectedCurrentTurn, newClientID string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if client exists in room (O(1) lookup)
	client := r.Clients[newClientID]
	if client == nil {
		return false // Client not in room
	}

	// Optimistic concurrency check: validate state matches
	if r.CurrentTurn != expectedCurrentTurn {
		return false // State mismatch
	}

	// End current turn if one is active (calculate duration and add to client's total)
	if r.CurrentTurn != "" && r.TurnStartTime != nil {
		r.endCurrentTurnLocked()
	}

	// Set the new turn
	r.CurrentTurn = newClientID
	now := time.Now().UnixNano()
	r.TurnStartTime = &now

	return true
}

// ClearCurrentTurn clears the current turn and adds duration to client's total
func (r *Room) ClearCurrentTurn() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.CurrentTurn != "" && r.TurnStartTime != nil {
		r.endCurrentTurnLocked()
	}
	r.CurrentTurn = ""
	r.TurnStartTime = nil
}

// endCurrentTurnLocked calculates the duration of the current turn and adds it to the client's total
// MUST be called with r.mu.Lock() held
func (r *Room) endCurrentTurnLocked() {
	if r.CurrentTurn == "" || r.TurnStartTime == nil {
		return
	}

	now := time.Now().UnixNano()
	durationMs := (now - *r.TurnStartTime) / int64(time.Millisecond)

	// Direct lookup - O(1)
	client := r.Clients[r.CurrentTurn]
	if client != nil {
		client.TotalTurnTime += durationMs
	}
}

// RemoveClient removes a client from the room (thread-safe)
// Returns true if the removed client had the current turn
func (r *Room) RemoveClient(clientID string) (bool, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if client exists in room (O(1) lookup)
	client := r.Clients[clientID]
	if client == nil {
		return false, false
	}

	// Check if this client had the current turn
	hadCurrentTurn := r.CurrentTurn == clientID

	if hadCurrentTurn && r.TurnStartTime != nil {
		// End their turn (calculate duration)
		r.endCurrentTurnLocked()
		r.CurrentTurn = ""
		r.TurnStartTime = nil
	}

	// Direct delete - O(1)
	delete(r.Clients, clientID)

	isEmpty := len(r.Clients) == 0
	return hadCurrentTurn, isEmpty
}
