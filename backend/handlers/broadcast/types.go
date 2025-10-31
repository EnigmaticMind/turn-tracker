package broadcast

import "encoding/json"

// BroadcastData is the data structure for broadcast messages
type BroadcastData struct {
	RoomID  string          `json:"room_id"`
	Payload json.RawMessage `json:"payload"`
}

// BroadcastReceivedData is the data structure for received broadcast messages
type BroadcastReceivedData struct {
	RoomID  string          `json:"room_id"`
	From    string          `json:"from"`
	Payload json.RawMessage `json:"payload"`
}
