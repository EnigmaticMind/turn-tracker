package core

type PeerInfo struct {
	ClientID    string `json:"client_id"`
	DisplayName string `json:"display_name"`
	Color       string `json:"color"`
}

type Room struct {
	ID        string
	Clients   map[*Client]bool
	CreatedBy string // clientID of the room creator
}

// NewRoom creates a new room
func NewRoom(id string) *Room {
	return &Room{
		ID:      id,
		Clients: make(map[*Client]bool),
	}
}

// ListPeerIDs returns all client IDs in the room (kept for backward compatibility)
func (r *Room) ListPeerIDs() []string {
	peers := make([]string, 0, len(r.Clients))
	for client := range r.Clients {
		peers = append(peers, client.ClientID)
	}
	return peers
}

// ListPeerInfo returns all peer information in the room
func (r *Room) ListPeerInfo() []PeerInfo {
	peers := make([]PeerInfo, 0, len(r.Clients))
	for client := range r.Clients {
		peers = append(peers, PeerInfo{
			ClientID:    client.ClientID,
			DisplayName: client.DisplayName,
			Color:       client.Color,
		})
	}
	return peers
}
