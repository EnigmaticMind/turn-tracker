package createroom

import (
	"encoding/json"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// NewRoomCreatedMessage creates a room_created message
func NewRoomCreatedMessage(roomID string, peers []core.PeerInfo) ([]byte, error) {
	data := RoomCreatedData{
		RoomID: roomID,
		Peers:  peers,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := types.Message{
		Type: "room_created",
		Data: dataJSON,
	}
	return json.Marshal(msg)
}
