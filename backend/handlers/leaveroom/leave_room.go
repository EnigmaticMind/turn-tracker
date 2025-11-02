package leaveroom

import (
	"log"
	"strings"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// HandleLeaveRoom handles a client leaving a room intentionally
func HandleLeaveRoom(hub *core.Hub, client *core.Client, roomID string) {
	// Normalize to uppercase for consistency
	roomID = strings.ToUpper(roomID)

	// Validate roomID matches client's current room
	if client.RoomID == "" {
		errorMsg, _ := types.NewErrorMessage("Not in a room")
		client.SafeSend(errorMsg)
		return
	}

	if client.RoomID != roomID {
		errorMsg, _ := types.NewErrorMessage("Room ID mismatch")
		client.SafeSend(errorMsg)
		return
	}

	// Get room before removing client
	room := hub.GetRoom(roomID)
	if room == nil {
		errorMsg, _ := types.NewErrorMessage("Room not found")
		client.SafeSend(errorMsg)
		return
	}

	// Remove client from room using centralized helper
	// This handles turn cleanup and notifications automatically
	// Don't clear client.RoomID - let handleUnregister handle the disconnect normally

	// Note: We don't delete empty rooms here - scheduled cleanup will handle it
	// This allows time for clients to reconnect and preserves their information

	// Note: We intentionally don't save user information in disconnectedClients,
	// this is an intentional exit of the room.

	hub.RemoveClientFromRoom(roomID, client.ClientID, "intentional leave")

	log.Printf("Client %s left room %s", client.ClientID, roomID)
}
