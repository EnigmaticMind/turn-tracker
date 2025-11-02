package joinroom

import (
	"encoding/json"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// NewRoomJoinedMessage creates a room_joined message
func NewRoomJoinedMessage(roomID, yourClientID string, peers []core.PeerInfo, currentTurn core.PeerInfo) ([]byte, error) {
	var currentTurnPtr *core.PeerInfo
	// If currentTurn is not empty (has a ClientID), use it; otherwise set to nil for null in JSON
	if currentTurn.ClientID != "" {
		currentTurnPtr = &currentTurn
	}

	data := RoomJoinedData{
		RoomID:       roomID,
		YourClientID: yourClientID,
		Peers:        peers,
		CurrentTurn:  currentTurnPtr,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := types.Message{
		Type: "room_joined",
		Data: dataJSON,
	}
	// Marshal message - json.Marshal allocates its own buffer
	marshaled, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	// Copy into pooled buffer for reuse
	return core.CopyToPooledBuffer(marshaled), nil
}

// NewPlayerJoinedMessage creates a player_joined message
func NewPlayerJoinedMessage(roomID, peerID, displayName, color string, totalTurnTime int64) ([]byte, error) {
	data := PlayerJoinedData{
		RoomID:        roomID,
		PeerID:        peerID,
		DisplayName:   displayName,
		Color:         color,
		TotalTurnTime: totalTurnTime,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := types.Message{
		Type: "player_joined",
		Data: dataJSON,
	}
	// Marshal message - json.Marshal allocates its own buffer
	marshaled, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	// Copy into pooled buffer for reuse
	return core.CopyToPooledBuffer(marshaled), nil
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
	// Marshal message - json.Marshal allocates its own buffer
	marshaled, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}
	// Copy into pooled buffer for reuse
	return core.CopyToPooledBuffer(marshaled), nil
}
