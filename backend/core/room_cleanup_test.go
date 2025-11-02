package core

import (
	"fmt"
	"testing"
	"time"
)

// Helper function to add rooms safely in tests
func addRoomForTest(hub *Hub, roomID string, room *Room) {
	hub.mu.Lock()
	defer hub.mu.Unlock()
	hub.rooms[roomID] = room
}

// TestCleanupAbandonedRooms wraps all room cleanup tests
// This allows running all tests together or individually in the IDE
func TestCleanupAbandonedRooms(t *testing.T) {
	t.Run("DeletesOldRooms", func(t *testing.T) {
		hub := NewHub()
		now := time.Now()

		// Create an old room (older than RoomAbandonTimeout)
		oldRoom := NewRoom("OLD123")
		oldRoom.CreatedAt = now.Add(-RoomAbandonTimeout - 1*time.Hour) // 13 hours ago
		addRoomForTest(hub, "OLD123", oldRoom)

		// Create a new room (newer than RoomAbandonTimeout)
		newRoom := NewRoom("NEW456")
		newRoom.CreatedAt = now.Add(-1 * time.Hour) // 1 hour ago
		addRoomForTest(hub, "NEW456", newRoom)

		// Verify both rooms exist
		if hub.GetRoom("OLD123") == nil {
			t.Fatal("Old room should exist before cleanup")
		}
		if hub.GetRoom("NEW456") == nil {
			t.Fatal("New room should exist before cleanup")
		}

		// Run cleanup
		hub.cleanupAbandonedRooms()

		// Old room should be deleted
		if hub.GetRoom("OLD123") != nil {
			t.Error("Old room should have been deleted")
		}

		// New room should still exist
		if hub.GetRoom("NEW456") == nil {
			t.Error("New room should not have been deleted")
		}
	})

	t.Run("SkipsZeroCreatedAt", func(t *testing.T) {
		hub := NewHub()

		// Create a room with zero CreatedAt
		zeroRoom := NewRoom("ZERO789")
		zeroRoom.CreatedAt = time.Time{} // Zero time
		addRoomForTest(hub, "ZERO789", zeroRoom)

		// Run cleanup
		hub.cleanupAbandonedRooms()

		// Room with zero CreatedAt should still exist (skipped)
		if hub.GetRoom("ZERO789") == nil {
			t.Error("Room with zero CreatedAt should be skipped, not deleted")
		}
	})

	t.Run("DeletesOldRoomsWithClients", func(t *testing.T) {
		hub := NewHub()
		now := time.Now()

		// Create an old room with clients
		oldRoomWithClients := NewRoom("OLDWITHCLIENTS")
		oldRoomWithClients.CreatedAt = now.Add(-RoomAbandonTimeout - 1*time.Hour) // 13 hours ago

		// Create a mock client (we need at least something in Clients map)
		mockClient := &Client{
			ClientID: "client1",
			RoomID:   "OLDWITHCLIENTS",
			Send:     make(chan []byte, 256), // Required field
		}
		oldRoomWithClients.Clients[mockClient.ClientID] = mockClient

		addRoomForTest(hub, "OLDWITHCLIENTS", oldRoomWithClients)

		// Verify room exists
		if hub.GetRoom("OLDWITHCLIENTS") == nil {
			t.Fatal("Old room with clients should exist before cleanup")
		}

		// Run cleanup - should still delete even with clients (per requirement)
		hub.cleanupAbandonedRooms()

		// Room should be deleted despite having clients
		if hub.GetRoom("OLDWITHCLIENTS") != nil {
			t.Error("Old room with clients should have been deleted (age-based cleanup)")
		}
	})

	t.Run("HandlesEmptyRoomsMap", func(t *testing.T) {
		hub := NewHub()

		// Verify it's empty by checking GetRoom returns nil for any ID
		if hub.GetRoom("NONEXISTENT") != nil {
			t.Error("Expected empty rooms map")
		}

		// Run cleanup on empty hub
		hub.cleanupAbandonedRooms()

		// Should not panic or error - verify still empty
		if hub.GetRoom("NONEXISTENT") != nil {
			t.Error("Expected empty rooms map after cleanup")
		}
	})

	t.Run("DeletesMultipleOldRooms", func(t *testing.T) {
		hub := NewHub()
		now := time.Now()

		// Create multiple old rooms
		for i := 0; i < 5; i++ {
			roomID := fmt.Sprintf("OLD%d", i)
			room := NewRoom(roomID)
			room.CreatedAt = now.Add(-RoomAbandonTimeout - time.Duration(i+1)*time.Hour)
			addRoomForTest(hub, roomID, room)
		}

		// Create some new rooms
		for i := 0; i < 3; i++ {
			roomID := fmt.Sprintf("NEW%d", i)
			room := NewRoom(roomID)
			room.CreatedAt = now.Add(-time.Duration(i+1) * time.Hour)
			addRoomForTest(hub, roomID, room)
		}

		// Run cleanup
		hub.cleanupAbandonedRooms()

		// All old rooms should be deleted
		for i := 0; i < 5; i++ {
			roomID := fmt.Sprintf("OLD%d", i)
			if hub.GetRoom(roomID) != nil {
				t.Errorf("Old room %s should have been deleted", roomID)
			}
		}

		// All new rooms should still exist
		for i := 0; i < 3; i++ {
			roomID := fmt.Sprintf("NEW%d", i)
			if hub.GetRoom(roomID) == nil {
				t.Errorf("New room %s should not have been deleted", roomID)
			}
		}
	})

	t.Run("ExactTimeoutBoundary", func(t *testing.T) {
		hub := NewHub()

		// Capture time once to use for both rooms
		baseTime := time.Now()

		// Room at exactly timeout minus small buffer to account for timing differences
		// This ensures the room is NOT deleted (age must be > timeout, not >=)
		exactRoom := NewRoom("EXACT")
		exactRoom.CreatedAt = baseTime.Add(-RoomAbandonTimeout + 1*time.Millisecond) // Slightly less than timeout
		addRoomForTest(hub, "EXACT", exactRoom)

		// Room just older than timeout (should be deleted)
		justOlderRoom := NewRoom("JUSTOLDER")
		justOlderRoom.CreatedAt = baseTime.Add(-RoomAbandonTimeout - 1*time.Minute)
		addRoomForTest(hub, "JUSTOLDER", justOlderRoom)

		// Run cleanup
		hub.cleanupAbandonedRooms()

		// Exact timeout room should NOT be deleted (age > timeout, not >=)
		if hub.GetRoom("EXACT") == nil {
			t.Error("Room exactly at timeout should not be deleted (boundary condition)")
		}

		// Just older room should be deleted
		if hub.GetRoom("JUSTOLDER") != nil {
			t.Error("Room just older than timeout should be deleted")
		}
	})
}
