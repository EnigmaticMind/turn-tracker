package main

import (
	"log"
	"net/http"

	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/joinroom"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

func serveWS(hub *core.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Check connection limit before creating client
	if !hub.TryRegister() {
		conn.Close()
		log.Printf("Connection rejected: maximum connections reached")
		return
	}

	// Create client without room assignment (will be assigned via create_room or join_room messages)
	client := &core.Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 8), // Reduced from 256 to 8 to limit memory per client
		ClientID: core.GenerateClientID(),
		RoomID:   "", // Will be set when room is created/joined
	}

	// Set message handler
	client.MessageHandler = messageRouter

	client.Hub.Register <- client

	// Allow collection of memory referenced by the caller by doing all work in new goroutines
	go client.WritePump()
	go client.ReadPump()
}

func main() {
	hub := core.NewHub()

	// Set up callback for player left notifications
	hub.OnPlayerLeft = func(roomID, clientID string, _ []byte) {
		playerLeftMsg, err := joinroom.NewPlayerLeftMessage(roomID, clientID)
		if err == nil {
			hub.BroadcastToRoom(roomID, playerLeftMsg)
		}
	}

	go hub.Run()

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})

	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
