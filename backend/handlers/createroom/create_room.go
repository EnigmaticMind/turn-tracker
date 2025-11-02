package createroom

import (
	"log"
	"turn-tracker/backend/core"
	"turn-tracker/backend/helpers"
	"turn-tracker/backend/types"
)

// HandleCreateRoom handles explicit room creation
// If roomID is empty, generates a new game ID
func HandleCreateRoom(hub *core.Hub, client *core.Client, roomID, displayName, color string) {
	// Initialize client profile (generates random if not provided)
	core.InitializeClientProfile(client, displayName, color)

	// Generate game ID if not provided or invalid
	if roomID == "" || !helpers.IsValidGameID(roomID) {
		// Generate unique game ID (with collision checking)
		for {
			roomID = helpers.GenerateGameID()
			if !hub.RoomExists(roomID) {
				break
			}
			// Collision detected, try again (extremely rare)
			log.Printf("Game ID collision detected, regenerating: %s", roomID)
		}
	} else {
		// Check if room already exists when ID is provided
		if hub.RoomExists(roomID) {
			errorMsg, _ := types.NewErrorMessage("Room already exists")
			client.SafeSend(errorMsg)
			return
		}
	}

	// Create room
	room := core.NewRoom(roomID)
	room.CreatedBy = client.ClientID
	room.Clients[client.ClientID] = client
	hub.AddRoom(roomID, room)

	// Update client's room ID
	client.RoomID = roomID

	// Send room_created message with peer info and current turn (if any)
	peers := room.ListPeerInfo()
	currentTurn := room.GetCurrentTurnInfo()
	response, err := NewRoomCreatedMessage(roomID, client.ClientID, peers, currentTurn)
	if err != nil {
		log.Printf("Error creating room_created message: %v", err)
		return
	}
	client.SafeSend(response)

	log.Printf("Room created: %s by client %s (%s)", roomID, client.ClientID, client.DisplayName)
}
