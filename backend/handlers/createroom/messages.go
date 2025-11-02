package createroom

import (
	"encoding/json"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// NewRoomCreatedMessage creates a room_created message
func NewRoomCreatedMessage(roomID, yourClientID string, peers []core.PeerInfo, currentTurn core.PeerInfo) ([]byte, error) {
	var currentTurnPtr *core.PeerInfo
	// If currentTurn is not empty (has a ClientID), use it; otherwise set to nil for null in JSON
	if currentTurn.ClientID != "" {
		currentTurnPtr = &currentTurn
	}

	data := RoomCreatedData{
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
		Type: "room_created",
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
