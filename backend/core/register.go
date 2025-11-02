package core

import (
	"log"
)

// handleRegister handles client registration
func (h *Hub) handleRegister(client *Client) {
	// Check if this is a reconnection with existing clientID
	if client.ClientID != "" {
		h.disconnectedMu.RLock()
		disconnected, exists := h.disconnectedClients[client.ClientID]
		h.disconnectedMu.RUnlock()

		if exists {
			// Restore client data from disconnected clients map
			client.DisplayName = disconnected.DisplayName
			client.Color = disconnected.Color
			client.TotalTurnTime = disconnected.TotalTurnTime
			// Remove from disconnectedClients (will be re-added if they disconnect again)
			h.disconnectedMu.Lock()
			delete(h.disconnectedClients, client.ClientID)
			h.disconnectedMu.Unlock()
			log.Printf("Client reconnected: %s (restored data)", client.ClientID)
		} else {
			// New client with provided ID, validate format
			if !isValidClientID(client.ClientID) {
				client.ClientID = GenerateClientID()
				log.Printf("Invalid clientID format, generated new: %s", client.ClientID)
			}
		}
	} else {
		// Generate new client ID
		client.ClientID = GenerateClientID()
	}
	h.clients[client] = true

	var remoteAddr string
	if client.Conn != nil {
		remoteAddr = client.Conn.RemoteAddr().String()
	} else {
		remoteAddr = "<no-connection>"
	}
	log.Printf("Client registered: %s (ID: %s)", remoteAddr, client.ClientID)
	// Room assignment happens via create_room or join_room messages
}

// isValidClientID validates that a clientID is exactly 16 hex characters
func isValidClientID(clientID string) bool {
	if len(clientID) != 16 {
		return false
	}
	for _, c := range clientID {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) {
			return false
		}
	}
	return true
}
