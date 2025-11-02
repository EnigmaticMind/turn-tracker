package core

import (
	"fmt"
	"testing"
	"time"
)

// Helper function to add disconnected clients safely in tests
func addDisconnectedClientForTest(hub *Hub, clientID string, client *DisconnectedClient) {
	hub.disconnectedMu.Lock()
	defer hub.disconnectedMu.Unlock()
	hub.disconnectedClients[clientID] = client
}

// TestCleanupDisconnectedClients wraps all disconnected client cleanup tests
// This allows running all tests together or individually in the IDE
func TestCleanupDisconnectedClients(t *testing.T) {
	t.Run("DeletesExpiredClients", func(t *testing.T) {
		hub := NewHub()
		now := time.Now()

		// Create an expired client (older than DisconnectedClientTTL)
		expiredClient := &DisconnectedClient{
			ClientID:       "expired1",
			DisplayName:    "Expired User",
			Color:          "#FF0000",
			LastRoomID:     "ROOM123",
			DisconnectedAt: now.Add(-DisconnectedClientTTL - 1*time.Minute), // 6 minutes ago
		}
		addDisconnectedClientForTest(hub, "expired1", expiredClient)

		// Create a non-expired client (newer than DisconnectedClientTTL)
		activeClient := &DisconnectedClient{
			ClientID:       "active1",
			DisplayName:    "Active User",
			Color:          "#00FF00",
			LastRoomID:     "ROOM456",
			DisconnectedAt: now.Add(-2 * time.Minute), // 2 minutes ago
		}
		addDisconnectedClientForTest(hub, "active1", activeClient)

		// Verify both clients exist
		hub.disconnectedMu.RLock()
		_, expiredExists := hub.disconnectedClients["expired1"]
		_, activeExists := hub.disconnectedClients["active1"]
		hub.disconnectedMu.RUnlock()

		if !expiredExists {
			t.Fatal("Expired client should exist before cleanup")
		}
		if !activeExists {
			t.Fatal("Active client should exist before cleanup")
		}

		// Run cleanup
		hub.cleanupDisconnectedClients()

		// Expired client should be deleted
		hub.disconnectedMu.RLock()
		_, expiredExists = hub.disconnectedClients["expired1"]
		_, activeExists = hub.disconnectedClients["active1"]
		hub.disconnectedMu.RUnlock()

		if expiredExists {
			t.Error("Expired client should have been deleted")
		}
		if !activeExists {
			t.Error("Active client should not have been deleted")
		}
	})

	t.Run("HandlesEmptyMap", func(t *testing.T) {
		hub := NewHub()

		// Run cleanup on empty hub
		hub.cleanupDisconnectedClients()

		// Should not panic or error
		hub.disconnectedMu.RLock()
		count := len(hub.disconnectedClients)
		hub.disconnectedMu.RUnlock()

		if count != 0 {
			t.Errorf("Expected empty disconnected clients map, got %d clients", count)
		}
	})

	t.Run("DeletesMultipleExpiredClients", func(t *testing.T) {
		hub := NewHub()
		now := time.Now()

		// Create multiple expired clients
		for i := 0; i < 5; i++ {
			clientID := fmt.Sprintf("expired%d", i)
			client := &DisconnectedClient{
				ClientID:       clientID,
				DisplayName:    fmt.Sprintf("User %d", i),
				DisconnectedAt: now.Add(-DisconnectedClientTTL - time.Duration(i+1)*time.Minute),
			}
			addDisconnectedClientForTest(hub, clientID, client)
		}

		// Create some non-expired clients
		for i := 0; i < 3; i++ {
			clientID := fmt.Sprintf("active%d", i)
			client := &DisconnectedClient{
				ClientID:       clientID,
				DisplayName:    fmt.Sprintf("User %d", i),
				DisconnectedAt: now.Add(-time.Duration(i+1) * time.Minute),
			}
			addDisconnectedClientForTest(hub, clientID, client)
		}

		// Run cleanup
		hub.cleanupDisconnectedClients()

		// All expired clients should be deleted
		hub.disconnectedMu.RLock()
		for i := 0; i < 5; i++ {
			clientID := fmt.Sprintf("expired%d", i)
			if _, exists := hub.disconnectedClients[clientID]; exists {
				t.Errorf("Expired client %s should have been deleted", clientID)
			}
		}

		// All non-expired clients should still exist
		for i := 0; i < 3; i++ {
			clientID := fmt.Sprintf("active%d", i)
			if _, exists := hub.disconnectedClients[clientID]; !exists {
				t.Errorf("Active client %s should not have been deleted", clientID)
			}
		}
		hub.disconnectedMu.RUnlock()
	})

	t.Run("ExactTimeoutBoundary", func(t *testing.T) {
		hub := NewHub()

		// Capture time once to use for both clients
		baseTime := time.Now()

		// Client exactly at timeout (should NOT be deleted - needs to be older)
		// Subtract a small buffer (1ms) to account for timing differences between test setup and cleanup
		exactClient := &DisconnectedClient{
			ClientID:       "exact",
			DisconnectedAt: baseTime.Add(-DisconnectedClientTTL + 1*time.Millisecond), // Slightly less than TTL
		}
		addDisconnectedClientForTest(hub, "exact", exactClient)

		// Client just older than timeout (should be deleted)
		justOlderClient := &DisconnectedClient{
			ClientID:       "justOlder",
			DisconnectedAt: baseTime.Add(-DisconnectedClientTTL - 1*time.Second), // 1 second older
		}
		addDisconnectedClientForTest(hub, "justOlder", justOlderClient)

		// Run cleanup
		hub.cleanupDisconnectedClients()

		// Exact timeout client should NOT be deleted (age > TTL, not >=)
		hub.disconnectedMu.RLock()
		_, exactExists := hub.disconnectedClients["exact"]
		_, justOlderExists := hub.disconnectedClients["justOlder"]
		hub.disconnectedMu.RUnlock()

		if !exactExists {
			t.Error("Client exactly at timeout should not be deleted (boundary condition)")
		}
		if justOlderExists {
			t.Error("Client just older than timeout should be deleted")
		}
	})
}
