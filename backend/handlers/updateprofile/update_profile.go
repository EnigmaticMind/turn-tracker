package updateprofile

import (
	"log"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// HandleUpdateProfile handles updating a user's profile (display name and/or color)
func HandleUpdateProfile(hub *core.Hub, client *core.Client, displayName, color string) {
	// Check if client is in a room
	if client.RoomID == "" {
		errorMsg, _ := types.NewErrorMessage("Not in a room")
		client.Send <- errorMsg
		return
	}

	// Update display name if provided
	if displayName != "" {
		client.DisplayName = displayName
	}

	// Update color if provided
	if color != "" {
		client.Color = color
	}

	// Broadcast profile update to other players in room
	profileUpdatedMsg, err := NewProfileUpdatedMessage(
		client.RoomID,
		client.ClientID,
		client.DisplayName,
		client.Color,
	)
	if err != nil {
		log.Printf("Error creating profile_updated message: %v", err)
		return
	}
	hub.BroadcastToRoomExcept(client.RoomID, client, profileUpdatedMsg)

	log.Printf("Client %s updated profile in room %s", client.ClientID, client.RoomID)
}
