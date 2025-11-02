package main

import (
	"encoding/json"
	"log"
	"strings"
	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/createroom"
	"turn-tracker/backend/handlers/joinroom"
	"turn-tracker/backend/handlers/leaveroom"
	"turn-tracker/backend/handlers/startturn"
	"turn-tracker/backend/handlers/updateprofile"
	"turn-tracker/backend/types"
)

// messageRouter routes incoming messages to the appropriate handler
// Fast path: msg.Type is already extracted, just route to handler
func messageRouter(hub *core.Hub, client *core.Client, msg *types.Message) {
	switch msg.Type {
	case "create_room":
		var data createroom.CreateRoomData
		// RoomID is optional, so we can unmarshal even if it's missing
		if unmarshalMessageData(msg, &data, "create_room", client) {
			// Normalize to uppercase for consistency
			roomID := strings.ToUpper(data.RoomID)
			createroom.HandleCreateRoom(hub, client, roomID, data.DisplayName, data.Color)
		}

	case "join_room":
		var data joinroom.JoinRoomData
		if unmarshalMessageData(msg, &data, "join_room", client) {
			// Normalize to uppercase for consistency
			roomID := strings.ToUpper(data.RoomID)
			joinroom.HandleJoinRoom(hub, client, roomID, data.DisplayName, data.Color)
		}

	case "leave_room":
		var data leaveroom.LeaveRoomData
		if unmarshalMessageData(msg, &data, "leave_room", client) {
			// Normalize to uppercase for consistency
			roomID := strings.ToUpper(data.RoomID)
			leaveroom.HandleLeaveRoom(hub, client, roomID)
		}

	case "update_profile":
		var data updateprofile.UpdateProfileData
		if unmarshalMessageData(msg, &data, "update_profile", client) {
			updateprofile.HandleUpdateProfile(hub, client, data.DisplayName, data.Color)
		}

	case "start_turn":
		var data startturn.StartTurnData
		if unmarshalMessageData(msg, &data, "start_turn", client) {
			startturn.HandleStartTurn(hub, client, data.CurrentTurn, data.NewTurn)
		}

	default:
		errorMsg, err := types.NewUnknownMessageTypeError(msg.Type)
		if err != nil {
			log.Printf("Failed to create unknown message type error: %v", err)
			return
		}
		client.SafeSend(errorMsg)
	}
}

// unmarshalMessageData unmarshals message data or sends error and returns false
func unmarshalMessageData(msg *types.Message, data interface{}, messageType string, client *core.Client) bool {
	if err := json.Unmarshal(msg.Data, data); err != nil {
		log.Printf("Error unmarshaling %s message: %v", messageType, err)
		errorMsg, err := types.NewErrorMessage("Invalid " + messageType + " data")
		if err != nil {
			log.Printf("Failed to create error message: %v", err)
			return false
		}
		client.SafeSend(errorMsg)
		return false
	}
	return true
}
