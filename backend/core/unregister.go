package core

import (
	"log"
	"time"
)

// handleUnregister handles client unregistration
func (h *Hub) handleUnregister(client *Client) {
	// Early return if client not registered (avoid unnecessary work)
	if _, ok := h.clients[client]; !ok {
		return
	}

	delete(h.clients, client)
	close(client.Send)

	// Cancel context to signal goroutines to stop
	if client.Cancel != nil {
		client.Cancel()
	}

	h.UnregisterConnection() // Decrement connection counter

	// Decrement IP connection count
	if client.IP != "" {
		h.UnregisterIP(client.IP)
	}

	roomID := client.RoomID
	clientID := client.ClientID

	// Save client data to disconnectedClients (only if needed for reconnection)
	// Only save if client was in a room (has roomID) and has profile data
	if roomID != "" && clientID != "" && (client.DisplayName != "" || client.Color != "") {
		disconnected := &DisconnectedClient{
			ClientID:       clientID,
			DisplayName:    client.DisplayName,
			Color:          client.Color,
			TotalTurnTime:  client.TotalTurnTime,
			LastRoomID:     roomID,
			DisconnectedAt: time.Now(),
		}

		h.disconnectedMu.Lock()
		h.disconnectedClients[clientID] = disconnected
		h.disconnectedMu.Unlock()
		log.Printf("Saved disconnected client data: %s", clientID)
	}

	var remoteAddr string
	if client.Conn != nil {
		remoteAddr = client.Conn.RemoteAddr().String()
	} else {
		remoteAddr = "<no-connection>"
	}

	// Get room reference (with read lock) - only if client was in a room
	if roomID == "" {
		log.Printf("Client unregistered: %s (ID: %s, no room)", remoteAddr, clientID)
		return
	}

	// Remove client from room using centralized helper
	// This handles turn cleanup and notifications automatically
	h.RemoveClientFromRoom(roomID, clientID, "disconnect")

	log.Printf("Client unregistered: %s (ID: %s)", remoteAddr, clientID)
}
