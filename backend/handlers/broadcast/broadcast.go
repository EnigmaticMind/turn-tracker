package broadcast

import (
	"encoding/json"
	"log"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// HandleBroadcast handles broadcasting a message to all peers in a room
func HandleBroadcast(hub *core.Hub, client *core.Client, roomID string, payload json.RawMessage) {
	room := hub.GetRoom(roomID)
	if room == nil {
		errorMsg, _ := types.NewErrorMessage("Room not found")
		client.Send <- errorMsg
		return
	}

	// Verify sender is in the room
	var sender *core.Client
	for c := range room.Clients {
		if c.ClientID == client.ClientID {
			sender = c
			break
		}
	}
	if sender == nil {
		errorMsg, _ := types.NewErrorMessage("Not a member of this room")
		client.Send <- errorMsg
		return
	}

	// Broadcast to all clients in room (including sender)
	broadcastMsg, err := NewBroadcastReceivedMessage(roomID, client.ClientID, payload)
	if err != nil {
		log.Printf("Error creating broadcast message: %v", err)
		return
	}
	hub.BroadcastToRoom(roomID, broadcastMsg)
}
