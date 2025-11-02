package core

import (
	"log"
	"time"
)

const (
	// DisconnectedClientTTL is how long to keep disconnected client data
	DisconnectedClientTTL = 5 * time.Minute
	// DisconnectedCleanupInterval is how often to clean up expired disconnected clients
	DisconnectedCleanupInterval = 5 * time.Minute
)

// StartDisconnectedCleanup starts a background goroutine to clean up expired disconnected clients
func (h *Hub) StartDisconnectedCleanup() {
	h.cleanupDone.Add(1)
	go func() {
		defer h.cleanupDone.Done()
		ticker := time.NewTicker(DisconnectedCleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-h.shutdownCtx.Done():
				log.Println("Disconnected client cleanup goroutine shutting down...")
				return
			case <-ticker.C:
				h.cleanupDisconnectedClients()
			}
		}
	}()
}

// cleanupDisconnectedClients removes expired disconnected client entries
func (h *Hub) cleanupDisconnectedClients() {
	now := time.Now()
	clientsToDelete := make([]string, 0)

	h.disconnectedMu.RLock()
	// Collect expired client IDs
	for clientID, client := range h.disconnectedClients {
		if now.Sub(client.DisconnectedAt) > DisconnectedClientTTL {
			clientsToDelete = append(clientsToDelete, clientID)
		}
	}
	h.disconnectedMu.RUnlock()

	// Early return if nothing to delete
	if len(clientsToDelete) == 0 {
		return
	}

	h.disconnectedMu.Lock()
	// Delete expired clients
	for _, clientID := range clientsToDelete {
		delete(h.disconnectedClients, clientID)
	}
	h.disconnectedMu.Unlock()

	// Log outside of lock to minimize lock time
	if len(clientsToDelete) > 0 {
		for _, clientID := range clientsToDelete {
			log.Printf("Removed expired disconnected client: %s", clientID)
		}
		log.Printf("Disconnected client cleanup: removed %d expired clients", len(clientsToDelete))
	}
}
