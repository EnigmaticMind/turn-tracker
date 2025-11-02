package createroom

import "turn-tracker/backend/core"

// CreateRoomData is the data structure for create_room messages
// RoomID is optional - if not provided or empty, backend will generate one
type CreateRoomData struct {
	RoomID      string `json:"room_id,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
	Color       string `json:"color,omitempty"`
}

// RoomCreatedData is the response data structure for room_created messages
type RoomCreatedData struct {
	RoomID      string          `json:"room_id"`
	YourClientID string         `json:"your_client_id"` // Client ID of the message recipient
	Peers       []core.PeerInfo `json:"peers"`
	CurrentTurn *core.PeerInfo  `json:"current_turn,omitempty"` // nil if no turn active
}
