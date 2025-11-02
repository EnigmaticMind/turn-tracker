package core

import (
	"sync"
	"sync/atomic"
	"testing"
)

// TestHub wraps all hub tests
// This allows running all tests together or individually in the IDE
func TestHub(t *testing.T) {
	t.Run("NewHub", func(t *testing.T) {
		hub := NewHub()

		if hub.rooms == nil {
			t.Error("Expected rooms map to be initialized")
		}
		if hub.clients == nil {
			t.Error("Expected clients map to be initialized")
		}
		if hub.disconnectedClients == nil {
			t.Error("Expected disconnectedClients map to be initialized")
		}
		if hub.Register == nil {
			t.Error("Expected Register channel to be initialized")
		}
		if hub.Unregister == nil {
			t.Error("Expected Unregister channel to be initialized")
		}
		if cap(hub.Register) != 100 {
			t.Errorf("Expected Register channel buffer size 100, got %d", cap(hub.Register))
		}
		if cap(hub.Unregister) != 100 {
			t.Errorf("Expected Unregister channel buffer size 100, got %d", cap(hub.Unregister))
		}
		if hub.currentConnections != 0 {
			t.Errorf("Expected currentConnections to be 0, got %d", hub.currentConnections)
		}
	})

	t.Run("TryRegister", func(t *testing.T) {
		t.Run("AllowsRegistrationUnderLimit", func(t *testing.T) {
			hub := NewHub()
			if !hub.TryRegister() {
				t.Error("Expected TryRegister to succeed when under limit")
			}
			if atomic.LoadInt32(&hub.currentConnections) != 1 {
				t.Errorf("Expected currentConnections to be 1, got %d", atomic.LoadInt32(&hub.currentConnections))
			}
		})

		t.Run("RejectsRegistrationAtLimit", func(t *testing.T) {
			hub := NewHub()
			// Set connections to max (using unsafe pointer cast or by calling TryRegister many times)
			// Since we can't directly set, we'll test by reaching the limit
			for i := 0; i < MaxConnections; i++ {
				hub.TryRegister()
			}

			if hub.TryRegister() {
				t.Error("Expected TryRegister to fail when at limit")
			}
			if atomic.LoadInt32(&hub.currentConnections) > MaxConnections {
				t.Errorf("Expected currentConnections not to exceed %d, got %d", MaxConnections, atomic.LoadInt32(&hub.currentConnections))
			}
		})

		t.Run("ConcurrentRegistrations", func(t *testing.T) {
			hub := NewHub()
			iterations := 100
			var wg sync.WaitGroup
			var successes int32

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if hub.TryRegister() {
						atomic.AddInt32(&successes, 1)
					}
				}()
			}

			wg.Wait()

			if successes != int32(iterations) {
				t.Errorf("Expected %d successful registrations, got %d", iterations, successes)
			}
			if atomic.LoadInt32(&hub.currentConnections) != int32(iterations) {
				t.Errorf("Expected currentConnections to be %d, got %d", iterations, atomic.LoadInt32(&hub.currentConnections))
			}
		})
	})

	t.Run("UnregisterConnection", func(t *testing.T) {
		t.Run("DecrementsCounter", func(t *testing.T) {
			hub := NewHub()
			hub.TryRegister()
			hub.TryRegister()

			if atomic.LoadInt32(&hub.currentConnections) != 2 {
				t.Errorf("Expected currentConnections to be 2, got %d", atomic.LoadInt32(&hub.currentConnections))
			}

			hub.UnregisterConnection()

			if atomic.LoadInt32(&hub.currentConnections) != 1 {
				t.Errorf("Expected currentConnections to be 1 after unregister, got %d", atomic.LoadInt32(&hub.currentConnections))
			}

			hub.UnregisterConnection()

			if atomic.LoadInt32(&hub.currentConnections) != 0 {
				t.Errorf("Expected currentConnections to be 0 after second unregister, got %d", atomic.LoadInt32(&hub.currentConnections))
			}
		})

		t.Run("ConcurrentUnregisters", func(t *testing.T) {
			hub := NewHub()
			iterations := 100

			// Register connections first
			for i := 0; i < iterations; i++ {
				hub.TryRegister()
			}

			var wg sync.WaitGroup
			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					hub.UnregisterConnection()
				}()
			}

			wg.Wait()

			if atomic.LoadInt32(&hub.currentConnections) != 0 {
				t.Errorf("Expected currentConnections to be 0 after all unregisters, got %d", atomic.LoadInt32(&hub.currentConnections))
			}
		})
	})

	t.Run("GetRoom", func(t *testing.T) {
		t.Run("ReturnsNilForNonExistentRoom", func(t *testing.T) {
			hub := NewHub()
			room := hub.GetRoom("NONEXISTENT")
			if room != nil {
				t.Error("Expected GetRoom to return nil for non-existent room")
			}
		})

		t.Run("ReturnsRoomWhenExists", func(t *testing.T) {
			hub := NewHub()
			testRoom := NewRoom("TEST123")
			hub.mu.Lock()
			hub.rooms["TEST123"] = testRoom
			hub.mu.Unlock()

			room := hub.GetRoom("TEST123")
			if room == nil {
				t.Error("Expected GetRoom to return room when it exists")
			}
			if room != testRoom {
				t.Error("Expected GetRoom to return the same room instance")
			}
		})

		t.Run("ThreadSafeReads", func(t *testing.T) {
			hub := NewHub()
			testRoom := NewRoom("TEST123")
			hub.mu.Lock()
			hub.rooms["TEST123"] = testRoom
			hub.mu.Unlock()

			var wg sync.WaitGroup
			iterations := 100

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					room := hub.GetRoom("TEST123")
					if room == nil {
						t.Error("Expected GetRoom to return room in concurrent reads")
					}
				}()
			}

			wg.Wait()
		})
	})

	t.Run("AddRoom", func(t *testing.T) {
		t.Run("AddsRoomToHub", func(t *testing.T) {
			hub := NewHub()
			testRoom := NewRoom("TEST123")

			hub.mu.Lock()
			hub.AddRoom("TEST123", testRoom)
			hub.mu.Unlock()

			if hub.GetRoom("TEST123") != testRoom {
				t.Error("Expected AddRoom to add room to hub")
			}
		})
	})

	t.Run("DeleteRoom", func(t *testing.T) {
		t.Run("DeletesRoomFromHub", func(t *testing.T) {
			hub := NewHub()
			testRoom := NewRoom("TEST123")
			hub.mu.Lock()
			hub.rooms["TEST123"] = testRoom
			hub.mu.Unlock()

			hub.DeleteRoom("TEST123")

			if hub.GetRoom("TEST123") != nil {
				t.Error("Expected DeleteRoom to remove room from hub")
			}
		})

		t.Run("HandlesNonExistentRoom", func(t *testing.T) {
			hub := NewHub()
			// Should not panic when deleting non-existent room
			hub.DeleteRoom("NONEXISTENT")
		})
	})

	t.Run("RoomExists", func(t *testing.T) {
		t.Run("ReturnsFalseForNonExistentRoom", func(t *testing.T) {
			hub := NewHub()
			if hub.RoomExists("NONEXISTENT") {
				t.Error("Expected RoomExists to return false for non-existent room")
			}
		})

		t.Run("ReturnsTrueForExistingRoom", func(t *testing.T) {
			hub := NewHub()
			testRoom := NewRoom("TEST123")
			hub.mu.Lock()
			hub.rooms["TEST123"] = testRoom
			hub.mu.Unlock()

			if !hub.RoomExists("TEST123") {
				t.Error("Expected RoomExists to return true for existing room")
			}
		})

		t.Run("ThreadSafe", func(t *testing.T) {
			hub := NewHub()
			testRoom := NewRoom("TEST123")
			hub.mu.Lock()
			hub.rooms["TEST123"] = testRoom
			hub.mu.Unlock()

			var wg sync.WaitGroup
			iterations := 100

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if !hub.RoomExists("TEST123") {
						t.Error("Expected RoomExists to return true in concurrent reads")
					}
				}()
			}

			wg.Wait()
		})
	})

	t.Run("HasDisconnectedClients", func(t *testing.T) {
		t.Run("ReturnsFalseWhenNoDisconnectedClients", func(t *testing.T) {
			hub := NewHub()
			if hub.HasDisconnectedClients("ROOM123") {
				t.Error("Expected HasDisconnectedClients to return false when no disconnected clients")
			}
		})

		t.Run("ReturnsTrueWhenDisconnectedClientExists", func(t *testing.T) {
			hub := NewHub()
			hub.disconnectedMu.Lock()
			hub.disconnectedClients["client1"] = &DisconnectedClient{
				ClientID:   "client1",
				LastRoomID: "ROOM123",
			}
			hub.disconnectedMu.Unlock()

			if !hub.HasDisconnectedClients("ROOM123") {
				t.Error("Expected HasDisconnectedClients to return true when disconnected client exists")
			}
		})

		t.Run("ReturnsFalseForDifferentRoom", func(t *testing.T) {
			hub := NewHub()
			hub.disconnectedMu.Lock()
			hub.disconnectedClients["client1"] = &DisconnectedClient{
				ClientID:   "client1",
				LastRoomID: "ROOM123",
			}
			hub.disconnectedMu.Unlock()

			if hub.HasDisconnectedClients("ROOM456") {
				t.Error("Expected HasDisconnectedClients to return false for different room")
			}
		})

		t.Run("HandlesMultipleDisconnectedClients", func(t *testing.T) {
			hub := NewHub()
			hub.disconnectedMu.Lock()
			hub.disconnectedClients["client1"] = &DisconnectedClient{
				ClientID:   "client1",
				LastRoomID: "ROOM123",
			}
			hub.disconnectedClients["client2"] = &DisconnectedClient{
				ClientID:   "client2",
				LastRoomID: "ROOM456",
			}
			hub.disconnectedClients["client3"] = &DisconnectedClient{
				ClientID:   "client3",
				LastRoomID: "ROOM123",
			}
			hub.disconnectedMu.Unlock()

			if !hub.HasDisconnectedClients("ROOM123") {
				t.Error("Expected HasDisconnectedClients to return true for ROOM123")
			}
			if !hub.HasDisconnectedClients("ROOM456") {
				t.Error("Expected HasDisconnectedClients to return true for ROOM456")
			}
			if hub.HasDisconnectedClients("ROOM789") {
				t.Error("Expected HasDisconnectedClients to return false for ROOM789")
			}
		})
	})
}
