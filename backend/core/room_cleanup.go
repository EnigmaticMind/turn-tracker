package core

import (
	"log"
	"time"
)

const (
	// Room cleanup constants
	RoomAbandonTimeout = 12 * time.Hour // Reduced from 24 hours
	CleanupInterval    = 2 * time.Hour  // Run less frequently (from 1 hour to 2 hours)
)

// StartRoomCleanup starts a background goroutine to clean up abandoned rooms
func (h *Hub) StartRoomCleanup() {
	h.cleanupDone.Add(1)
	go func() {
		defer h.cleanupDone.Done()
		ticker := time.NewTicker(CleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-h.shutdownCtx.Done():
				log.Println("Room cleanup goroutine shutting down...")
				return
			case <-ticker.C:
				h.cleanupAbandonedRooms()
			}
		}
	}()
}

func (h *Hub) cleanupAbandonedRooms() {
	now := time.Now()
	roomsToDelete := make([]string, 0)

	// First pass: read lock to identify rooms to delete
	h.mu.RLock()
	for roomID, room := range h.rooms {
		room.mu.RLock()
		// Check for uninitialized CreatedAt
		if room.CreatedAt.IsZero() {
			room.mu.RUnlock()
			log.Printf("Room cleanup: warning - room %s has zero CreatedAt, skipping", roomID)
			continue
		}
		age := now.Sub(room.CreatedAt)
		room.mu.RUnlock()

		// Collect rooms older than RoomAbandonTimeout for deletion
		if age > RoomAbandonTimeout {
			roomsToDelete = append(roomsToDelete, roomID)
		}
	}
	h.mu.RUnlock()

	// Second pass: write lock to actually delete rooms
	if len(roomsToDelete) == 0 {
		return
	}

	h.mu.Lock()
	deletedCount := 0
	for _, roomID := range roomsToDelete {
		// Double-check room still exists and verify age before deletion
		room, exists := h.rooms[roomID]
		if !exists {
			continue // Room was already deleted
		}

		room.mu.RLock()
		isZero := room.CreatedAt.IsZero()
		age := now.Sub(room.CreatedAt)
		clientCount := len(room.Clients)
		room.mu.RUnlock()

		if !isZero && age > RoomAbandonTimeout {
			delete(h.rooms, roomID)
			if clientCount > 0 {
				log.Printf("Room cleanup: deleted room %s with %d active clients (age: %v)", roomID, clientCount, age.Round(time.Minute))
			} else {
				log.Printf("Room cleanup: deleted room %s (age: %v)", roomID, age.Round(time.Minute))
			}
			deletedCount++
		}
	}
	h.mu.Unlock()

	if deletedCount > 0 {
		log.Printf("Room cleanup: deleted %d rooms older than %v", deletedCount, RoomAbandonTimeout)
	}
}
