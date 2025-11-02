package startturn

import "turn-tracker/backend/core"

// StartTurnData is the data structure for start_turn messages
type StartTurnData struct {
	CurrentTurn string `json:"current_turn"`       // Client's view of current turn (empty string if no turn)
	NewTurn     string `json:"new_turn,omitempty"` // Client ID to start turn for (empty means end turn)
}

// TurnChangedData is the data structure for turn_changed messages
type TurnChangedData struct {
	RoomID        string         `json:"room_id"`
	CurrentTurn   *core.PeerInfo `json:"current_turn"`    // nil if no turn active
	TurnStartTime *int64         `json:"turn_start_time"` // Unix timestamp in milliseconds when turn started (nil if no turn active)
	Sequence      uint64         `json:"sequence"`        // Sequence number to identify stale messages (higher = newer)
}
