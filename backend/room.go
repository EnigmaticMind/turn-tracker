package main

type Room struct {
	ID      string
	clients map[*Client]bool
	// Add room-specific data here (players, current turn, etc.)
}

func newRoom(id string) *Room {
	return &Room{
		ID:      id,
		clients: make(map[*Client]bool),
	}
}

