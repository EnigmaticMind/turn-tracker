package main

import (
	"encoding/json"
	"log"
)

type Hub struct {
	rooms map[string]*Room
	// Registered clients.
	clients map[*Client]bool
	// Inbound messages from the clients.
	broadcast chan []byte
	// Register requests from the clients.
	register chan *Client
	// Unregister requests from clients.
	unregister chan *Client
}

func newHub() *Hub {
	return &Hub{
		rooms:      make(map[string]*Room),
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

func (h *Hub) run() {
	for {
		select {
		case client := <-h.register:
			h.clients[client] = true
			log.Printf("Client registered: %s", client.conn.RemoteAddr())

			// Get or create room
			room := h.getOrCreateRoom(client.roomID)

			// Add client to room
			room.clients[client] = true

			// Broadcast player joined message
			message := map[string]interface{}{
				"type": "player_joined",
				"data": map[string]interface{}{
					"room": client.roomID,
				},
			}
			jsonData, _ := json.Marshal(message)
			h.broadcastToRoom(client.roomID, jsonData)

		case client := <-h.unregister:
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.send)

				// Remove client from room
				if room, exists := h.rooms[client.roomID]; exists {
					delete(room.clients, client)

					// Broadcast player left message
					message := map[string]interface{}{
						"type": "player_left",
						"data": map[string]interface{}{
							"room": client.roomID,
						},
					}
					jsonData, _ := json.Marshal(message)
					h.broadcastToRoom(client.roomID, jsonData)
				}

				log.Printf("Client unregistered: %s", client.conn.RemoteAddr())
			}

		case message := <-h.broadcast:
			for client := range h.clients {
				select {
				case client.send <- message:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

func (h *Hub) getOrCreateRoom(roomID string) *Room {
	if room, exists := h.rooms[roomID]; exists {
		return room
	}
	room := newRoom(roomID)
	h.rooms[roomID] = room
	return room
}

func (h *Hub) broadcastToRoom(roomID string, message []byte) {
	if room, exists := h.rooms[roomID]; exists {
		for client := range room.clients {
			select {
			case client.send <- message:
			default:
				close(client.send)
				delete(room.clients, client)
			}
		}
	}
}

