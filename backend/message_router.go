package main

import (
	"encoding/json"
	"strings"
	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/broadcast"
	"turn-tracker/backend/handlers/createroom"
	"turn-tracker/backend/handlers/joinroom"
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
		json.Unmarshal(msg.Data, &data) // Ignore errors, RoomID will be empty if missing
		createroom.HandleCreateRoom(hub, client, data.RoomID, data.DisplayName, data.Color)

	case "join_room":
		var data joinroom.JoinRoomData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			errorMsg, _ := types.NewErrorMessage("Invalid join_room data")
			client.Send <- errorMsg
			return
		}
		// Normalize to uppercase for consistency
		roomID := strings.ToUpper(data.RoomID)
		joinroom.HandleJoinRoom(hub, client, roomID, data.DisplayName, data.Color)

	case "broadcast":
		var data broadcast.BroadcastData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			errorMsg, _ := types.NewErrorMessage("Invalid broadcast data")
			client.Send <- errorMsg
			return
		}
		broadcast.HandleBroadcast(hub, client, data.RoomID, data.Payload)

	case "update_profile":
		var data updateprofile.UpdateProfileData
		if err := json.Unmarshal(msg.Data, &data); err != nil {
			errorMsg, _ := types.NewErrorMessage("Invalid update_profile data")
			client.Send <- errorMsg
			return
		}
		updateprofile.HandleUpdateProfile(hub, client, data.DisplayName, data.Color)

	default:
		errorMsg, _ := types.NewUnknownMessageTypeError(msg.Type)
		client.Send <- errorMsg
	}
}
