package startturn

import (
	"encoding/json"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// NewTurnChangedMessage creates a turn_changed message with a sequence number
func NewTurnChangedMessage(roomID string, currentTurn core.PeerInfo, turnStartTime int64, sequence uint64) ([]byte, error) {
	var currentTurnPtr *core.PeerInfo
	// If currentTurn is not empty (has a ClientID), use it; otherwise set to nil for null in JSON
	if currentTurn.ClientID != "" {
		currentTurnPtr = &currentTurn
	}

	var turnStartTimePtr *int64
	// If turnStartTime is not 0, use it; otherwise set to nil for null in JSON
	if turnStartTime != 0 {
		turnStartTimePtr = &turnStartTime
	}

	data := TurnChangedData{
		RoomID:        roomID,
		CurrentTurn:   currentTurnPtr,
		TurnStartTime: turnStartTimePtr,
		Sequence:      sequence,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := types.Message{
		Type: "turn_changed",
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
