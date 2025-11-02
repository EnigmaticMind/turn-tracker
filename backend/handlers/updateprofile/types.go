package updateprofile

// UpdateProfileData is the data structure for update_profile messages
type UpdateProfileData struct {
	DisplayName string `json:"display_name,omitempty"`
	Color       string `json:"color,omitempty"`
}

// ProfileUpdatedData is the data structure for profile_updated messages
type ProfileUpdatedData struct {
	RoomID        string `json:"room_id"`
	PeerID        string `json:"peer_id"`
	DisplayName   string `json:"display_name"`
	Color         string `json:"color"`
	TotalTurnTime int64  `json:"total_turn_time"` // Total time spent in turns (in milliseconds)
}
