package broadcast

import (
	"encoding/json"
	"turn-tracker/backend/types"
)

// NewBroadcastReceivedMessage creates a broadcast message
func NewBroadcastReceivedMessage(roomID, from string, payload json.RawMessage) ([]byte, error) {
	data := BroadcastReceivedData{
		RoomID:  roomID,
		From:    from,
		Payload: payload,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := types.Message{
		Type: "broadcast",
		Data: dataJSON,
	}
	return json.Marshal(msg)
}
