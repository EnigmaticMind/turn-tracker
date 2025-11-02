package core

// BroadcastToRoomExcept broadcasts a message to all clients in a room except the specified client
// If except is nil, broadcasts to all clients in the room
func (h *Hub) BroadcastToRoomExcept(roomID string, except *Client, message []byte) {
	h.mu.RLock()
	room, exists := h.rooms[roomID]
	h.mu.RUnlock()

	if !exists {
		return
	}

	room.mu.RLock()
	// Create a copy of clients to iterate over safely
	clients := make([]*Client, 0, len(room.Clients))
	for _, client := range room.Clients {
		if except == nil || client != except {
			clients = append(clients, client)
		}
	}
	room.mu.RUnlock()

	// Send to each client - let SafeSend handle closed channels gracefully
	for _, client := range clients {
		// Use SafeSend which handles closed channels and full channels properly
		if !client.SafeSend(message) {
			// Channel is full or closed - client might be dead
			// Check if client is still registered in hub (better check)
			h.mu.RLock()
			_, stillRegistered := h.clients[client]
			h.mu.RUnlock()

			if !stillRegistered {
				// Client was unregistered, safe to remove from room
				// Use centralized helper to remove and notify
				h.RemoveClientFromRoom(roomID, client.ClientID, "channel closed during broadcast")
			}
			// If still registered but channel is full, just skip
			// They might be slow, not dead
		}
	}
}

// BroadcastToRoom broadcasts a message to all clients in a room
func (h *Hub) BroadcastToRoom(roomID string, message []byte) {
	h.BroadcastToRoomExcept(roomID, nil, message)
}
