package updateprofile

import (
	"log"
	"strings"
	"turn-tracker/backend/core"
	"turn-tracker/backend/helpers"
	"turn-tracker/backend/types"
)

const (
	maxDisplayNameLength = 50
)

// HandleUpdateProfile handles updating a user's profile (display name and/or color)
func HandleUpdateProfile(hub *core.Hub, client *core.Client, displayName, color string) {
	// Check if client is in a room
	if client.RoomID == "" {
		errorMsg, _ := types.NewErrorMessage("Not in a room")
		client.SafeSend(errorMsg)
		return
	}

	// Early return if no updates needed
	if (displayName == "" || displayName == client.DisplayName) && (color == "" || color == client.Color) {
		// Nothing to update, but send confirmation to sender
		return
	}

	// Validate inputs
	if displayName != "" {
		displayName = strings.TrimSpace(displayName)
		if !helpers.IsValidDisplayName(displayName) {
			errorMsg, _ := types.NewErrorMessage("Invalid display name")
			client.SafeSend(errorMsg)
			return
		}
	}

	if color != "" {
		color = strings.ToUpper(strings.TrimSpace(color))
		if !helpers.IsValidHexColor(color) {
			errorMsg, _ := types.NewErrorMessage("Invalid color format (expected #RRGGBB)")
			client.SafeSend(errorMsg)
			return
		}
	}

	// Get room to verify it exists
	room := hub.GetRoom(client.RoomID)
	if room == nil {
		errorMsg, _ := types.NewErrorMessage("Room not found")
		client.SafeSend(errorMsg)
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
		client.TotalTurnTime,
	)
	if err != nil {
		log.Printf("Error creating profile_updated message: %v", err)
		return
	}
	hub.BroadcastToRoom(client.RoomID, profileUpdatedMsg)

	log.Printf("Client %s updated profile in room %s", client.ClientID, client.RoomID)
}
