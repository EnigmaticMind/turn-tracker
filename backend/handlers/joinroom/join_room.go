package joinroom

import (
	"log"

	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// HandleJoinRoom handles joining an existing room
func HandleJoinRoom(hub *core.Hub, client *core.Client, roomID, displayName, color string) {
	// Initialize client profile (generates random if not provided)
	core.InitializeClientProfile(client, displayName, color)

	// Validate game ID format first
	if !core.IsValidGameID(roomID) {
		errorMsg, _ := types.NewErrorMessage("Invalid game ID format")
		client.Send <- errorMsg
		return
	}

	// Check if room exists
	room := hub.GetRoom(roomID)
	if room == nil {
		errorMsg, _ := types.NewErrorMessage("Room not found")
		client.Send <- errorMsg
		return
	}

	// Add client to room
	room.Clients[client] = true
	client.RoomID = roomID

	// Send room_joined message to joiner with peer info
	peers := room.ListPeerInfo()
	response, err := NewRoomJoinedMessage(roomID, peers)
	if err != nil {
		log.Printf("Error creating room_joined message: %v", err)
		return
	}
	client.Send <- response

	// Notify other players in room about the new joiner (with profile info)
	playerJoinedMsg, err := NewPlayerJoinedMessage(roomID, client.ClientID, client.DisplayName, client.Color)
	if err != nil {
		log.Printf("Error creating player_joined message: %v", err)
		return
	}
	hub.BroadcastToRoomExcept(roomID, client, playerJoinedMsg)

	log.Printf("Client %s (%s) joined room %s", client.ClientID, client.DisplayName, roomID)
}
