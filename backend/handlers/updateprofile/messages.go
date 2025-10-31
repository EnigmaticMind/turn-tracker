package updateprofile

import (
	"encoding/json"
	"turn-tracker/backend/types"
)

// NewProfileUpdatedMessage creates a profile_updated message
func NewProfileUpdatedMessage(roomID, peerID, displayName, color string) ([]byte, error) {
	data := ProfileUpdatedData{
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
		Type: "profile_updated",
		Data: dataJSON,
	}
	return json.Marshal(msg)
}
