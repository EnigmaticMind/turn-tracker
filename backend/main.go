package main

import (
	"context"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/joinroom"
	"turn-tracker/backend/handlers/startturn"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

// Add helper function
func getClientIP(r *http.Request) string {
	// Check X-Forwarded-For (for reverse proxies)
	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			return strings.TrimSpace(ips[0])
		}
	}

	// Check X-Real-IP
	if xri := r.Header.Get("X-Real-IP"); xri != "" {
		return xri
	}

	// Fallback to RemoteAddr
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err != nil {
		return r.RemoteAddr
	}
	return host
}

func serveWS(hub *core.Hub, w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
		return
	}

	// Get client IP and check limits
	clientIP := getClientIP(r)
	if !hub.TryRegisterIP(clientIP) {
		conn.Close()
		log.Printf("Connection rejected from IP %s: connection limit reached", clientIP)
		return
	}

	// Try to get clientID from query parameter (for reconnection)
	clientID := r.URL.Query().Get("client_id")

	// Create client without room assignment (will be assigned via create_room or join_room messages)
	client := &core.Client{
		Hub:      hub,
		Conn:     conn,
		Send:     make(chan []byte, 32), // Reduced from 256 to 8 to limit memory per client
		ClientID: clientID,              // Will be validated/generated in hub.Register
		RoomID:   "",                    // Will be set when room is created/joined
		IP:       clientIP,              // Store IP for cleanup
	}

	client.Ctx, client.Cancel = context.WithCancel(context.Background())

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

	// Set up callback for turn ended (when player disconnects during their turn)
	hub.OnTurnEnded = func(roomID string) {
		room := hub.GetRoom(roomID)
		if room == nil {
			return
		}
		// Get sequence number for turn_changed message
		sequence := room.GetTurnSequence()
		turnChangedMsg, err := startturn.NewTurnChangedMessage(roomID, core.PeerInfo{}, 0, sequence)
		if err == nil {
			hub.BroadcastToRoom(roomID, turnChangedMsg)
		}
	}

	go hub.Run()

	// Start room cleanup goroutine
	hub.StartRoomCleanup()

	// Start disconnected client cleanup goroutine
	hub.StartDisconnectedCleanup()

	healthCheck := func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers to allow cross-origin requests
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight OPTIONS request
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		// Simple text response - no JSON encoding needed
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	}

	// Health check endpoint
	http.HandleFunc("/health", healthCheck)
	http.HandleFunc("/", healthCheck)

	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		serveWS(hub, w, r)
	})

	// Get port from environment variable (required for fly.io)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	// Create HTTP server with timeouts
	server := &http.Server{
		Addr:         ":" + port,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on :%s", port)
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed to start: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Stop accepting new connections
	if err := server.Shutdown(ctx); err != nil {
		log.Printf("Server forced to shutdown: %v", err)
	}

	// Shutdown hub and all its goroutines
	hub.Shutdown()

	log.Println("Server exited")
}
