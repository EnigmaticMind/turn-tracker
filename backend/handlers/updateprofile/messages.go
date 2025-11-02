package updateprofile

import (
	"encoding/json"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// NewProfileUpdatedMessage creates a profile_updated message
func NewProfileUpdatedMessage(roomID, peerID, displayName, color string, totalTurnTime int64) ([]byte, error) {
	data := ProfileUpdatedData{
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
		Type: "profile_updated",
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
