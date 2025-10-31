package joinroom

import (
	"encoding/json"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// NewRoomJoinedMessage creates a room_joined message
func NewRoomJoinedMessage(roomID string, peers []core.PeerInfo) ([]byte, error) {
	data := RoomJoinedData{
		RoomID: roomID,
		Peers:  peers,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := types.Message{
		Type: "room_joined",
		Data: dataJSON,
	}
	return json.Marshal(msg)
}

// NewPlayerJoinedMessage creates a player_joined message
func NewPlayerJoinedMessage(roomID, peerID, displayName, color string) ([]byte, error) {
	data := PlayerJoinedData{
		RoomID:      roomID,
		PeerID:      peerID,
		DisplayName: displayName,
		Color:       color,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := types.Message{
		Type: "player_joined",
		Data: dataJSON,
	}
	return json.Marshal(msg)
}

// NewPlayerLeftMessage creates a player_left message
func NewPlayerLeftMessage(roomID, peerID string) ([]byte, error) {
	data := PlayerLeftData{
		RoomID: roomID,
		PeerID: peerID,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := types.Message{
		Type: "player_left",
		Data: dataJSON,
	}
	return json.Marshal(msg)
}
