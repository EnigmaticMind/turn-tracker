package joinroom

import (
	"log"

	"turn-tracker/backend/core"
	"turn-tracker/backend/helpers"
	"turn-tracker/backend/types"
)

// HandleJoinRoom handles joining an existing room
func HandleJoinRoom(hub *core.Hub, client *core.Client, roomID, displayName, color string) {
	// Initialize client profile (generates random if not provided)
	core.InitializeClientProfile(client, displayName, color)

	// Validate game ID format first
	if !helpers.IsValidGameID(roomID) {
		errorMsg, _ := types.NewErrorMessage("Invalid game ID format")
		client.SafeSend(errorMsg)
		return
	}

	// Check if room exists
	room := hub.GetRoom(roomID)
	if room == nil {
		errorMsg, _ := types.NewErrorMessage("Room not found")
		client.SafeSend(errorMsg)
		return
	}

	// Check if client is already in another room
	if client.RoomID != "" && client.RoomID != roomID {
		oldRoomID := client.RoomID
		oldRoom := hub.GetRoom(oldRoomID)

		if oldRoom != nil {
			// Remove client from OLD room (not the new one) using centralized helper
			// This handles turn cleanup and notifications automatically
			hub.RemoveClientFromRoom(oldRoomID, client.ClientID, "moved to another room")
		} else {
			// Old room doesn't exist - client in inconsistent state
			log.Printf("Client %s had invalid RoomID %s (room not found), clearing it", client.ClientID, oldRoomID)
			client.RoomID = ""
		}
	}

	// Re-validate room still exists after potential leave operation
	room = hub.GetRoom(roomID)
	if room == nil {
		errorMsg, _ := types.NewErrorMessage("Room has been deleted")
		client.SafeSend(errorMsg)
		return
	}

	// Add client to room first (so they're included in peers list)
	if !room.AddClient(client) {
		errorMsg, _ := types.NewErrorMessage("Already in this room")
		client.SafeSend(errorMsg)
		return
	}

	// Update client's room ID
	client.RoomID = roomID

	// Now get peers list (includes the joining client)
	peers := room.ListPeerInfo()
	currentTurn := room.GetCurrentTurnInfo()
	response, err := NewRoomJoinedMessage(roomID, client.ClientID, peers, currentTurn)
	if err != nil {
		log.Printf("Error creating room_joined message: %v", err)
		// Rollback: remove client from room
		room.RemoveClient(client.ClientID)
		client.RoomID = ""
		errorMsg, _ := types.NewErrorMessage("Failed to create join message")
		client.SafeSend(errorMsg)
		return
	}

	playerJoinedMsg, err := NewPlayerJoinedMessage(roomID, client.ClientID, client.DisplayName, client.Color, client.TotalTurnTime)
	if err != nil {
		log.Printf("Error creating player_joined message: %v", err)
		// Rollback: remove client from room
		room.RemoveClient(client.ClientID)
		client.RoomID = ""
		errorMsg, _ := types.NewErrorMessage("Failed to create notification message")
		client.SafeSend(errorMsg)
		return
	}

	// Send messages
	client.SafeSend(response)
	hub.BroadcastToRoomExcept(roomID, client, playerJoinedMsg)
	log.Printf("Client %s (%s) joined room %s", client.ClientID, client.DisplayName, roomID)
}
