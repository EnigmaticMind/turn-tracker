package joinroom

import "turn-tracker/backend/core"

// JoinRoomData is the data structure for join_room messages
type JoinRoomData struct {
	RoomID      string `json:"room_id"`
	DisplayName string `json:"display_name,omitempty"`
	Color       string `json:"color,omitempty"`
}

// RoomJoinedData is the response data structure for room_joined messages
type RoomJoinedData struct {
	RoomID string          `json:"room_id"`
	Peers  []core.PeerInfo `json:"peers"`
}

// PlayerJoinedData is the data structure for player_joined messages
type PlayerJoinedData struct {
	RoomID      string `json:"room_id"`
	PeerID      string `json:"peer_id"`
	DisplayName string `json:"display_name"`
	Color       string `json:"color"`
}

// PlayerLeftData is the data structure for player_left messages
type PlayerLeftData struct {
	RoomID string `json:"room_id"`
	PeerID string `json:"peer_id"`
}
