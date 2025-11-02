package core

import (
	"testing"
	"time"
)

// createTestClient creates a minimal client for testing
func createTestClient(clientID, displayName, color string) *Client {
	return &Client{
		ClientID:      clientID,
		DisplayName:   displayName,
		Color:         color,
		TotalTurnTime: 0,
	}
}

func TestRoom(t *testing.T) {
	t.Run("NewRoom", func(t *testing.T) {
		roomID := "TEST123"
		room := NewRoom(roomID)

		if room.ID != roomID {
			t.Errorf("Expected room ID %s, got %s", roomID, room.ID)
		}

		if room.Clients == nil {
			t.Error("Expected Clients map to be initialized, got nil")
		}

		if len(room.Clients) != 0 {
			t.Errorf("Expected empty Clients map, got %d clients", len(room.Clients))
		}

		if room.CreatedAt.IsZero() {
			t.Error("Expected CreatedAt to be set, got zero time")
		}
	})

	t.Run("AddClient", func(t *testing.T) {
		room := NewRoom("TEST123")
		client1 := createTestClient("client1", "Alice", "#FF0000")

		// Add first client
		if !room.AddClient(client1) {
			t.Error("Expected AddClient to return true for new client")
		}

		if len(room.Clients) != 1 {
			t.Errorf("Expected 1 client in room, got %d", len(room.Clients))
		}

		if room.Clients["client1"] != client1 {
			t.Error("Expected client1 to be in room")
		}

		// Try to add same client again
		if room.AddClient(client1) {
			t.Error("Expected AddClient to return false for duplicate client")
		}

		if len(room.Clients) != 1 {
			t.Errorf("Expected still 1 client in room, got %d", len(room.Clients))
		}

		// Add second client
		client2 := createTestClient("client2", "Bob", "#00FF00")
		if !room.AddClient(client2) {
			t.Error("Expected AddClient to return true for second client")
		}

		if len(room.Clients) != 2 {
			t.Errorf("Expected 2 clients in room, got %d", len(room.Clients))
		}
	})

	t.Run("ListPeerIDs", func(t *testing.T) {
		t.Run("EmptyRoom", func(t *testing.T) {
			room := NewRoom("TEST123")
			peers := room.ListPeerIDs()

			if peers != nil {
				t.Errorf("Expected nil for empty room, got %v", peers)
			}
		})

		t.Run("WithClients", func(t *testing.T) {
			room := NewRoom("TEST123")
			client1 := createTestClient("client1", "Alice", "#FF0000")
			client2 := createTestClient("client2", "Bob", "#00FF00")

			room.AddClient(client1)
			room.AddClient(client2)

			peers := room.ListPeerIDs()

			if len(peers) != 2 {
				t.Errorf("Expected 2 peer IDs, got %d", len(peers))
			}

			// Check that both client IDs are present (order not guaranteed)
			peerMap := make(map[string]bool)
			for _, id := range peers {
				peerMap[id] = true
			}

			if !peerMap["client1"] {
				t.Error("Expected client1 in peer IDs")
			}
			if !peerMap["client2"] {
				t.Error("Expected client2 in peer IDs")
			}
		})
	})

	t.Run("ListPeerInfo", func(t *testing.T) {
		t.Run("EmptyRoom", func(t *testing.T) {
			room := NewRoom("TEST123")
			peers := room.ListPeerInfo()

			if len(peers) != 0 {
				t.Errorf("Expected empty peer list, got %d peers", len(peers))
			}
		})

		t.Run("WithClients", func(t *testing.T) {
			room := NewRoom("TEST123")
			client1 := createTestClient("client1", "Alice", "#FF0000")
			client1.TotalTurnTime = 5000
			client2 := createTestClient("client2", "Bob", "#00FF00")
			client2.TotalTurnTime = 3000

			room.AddClient(client1)
			room.AddClient(client2)

			peers := room.ListPeerInfo()

			if len(peers) != 2 {
				t.Errorf("Expected 2 peers, got %d", len(peers))
			}

			// Create a map for easy lookup
			peerMap := make(map[string]PeerInfo)
			for _, peer := range peers {
				peerMap[peer.ClientID] = peer
			}

			// Check client1
			peer1, ok := peerMap["client1"]
			if !ok {
				t.Error("Expected client1 in peer list")
			} else {
				if peer1.DisplayName != "Alice" {
					t.Errorf("Expected DisplayName 'Alice', got '%s'", peer1.DisplayName)
				}
				if peer1.Color != "#FF0000" {
					t.Errorf("Expected Color '#FF0000', got '%s'", peer1.Color)
				}
				if peer1.TotalTurnTime != 5000 {
					t.Errorf("Expected TotalTurnTime 5000, got %d", peer1.TotalTurnTime)
				}
			}

			// Check client2
			peer2, ok := peerMap["client2"]
			if !ok {
				t.Error("Expected client2 in peer list")
			} else {
				if peer2.DisplayName != "Bob" {
					t.Errorf("Expected DisplayName 'Bob', got '%s'", peer2.DisplayName)
				}
				if peer2.Color != "#00FF00" {
					t.Errorf("Expected Color '#00FF00', got '%s'", peer2.Color)
				}
				if peer2.TotalTurnTime != 3000 {
					t.Errorf("Expected TotalTurnTime 3000, got %d", peer2.TotalTurnTime)
				}
			}
		})
	})

	t.Run("GetCurrentTurnInfo", func(t *testing.T) {
		t.Run("NoTurn", func(t *testing.T) {
			room := NewRoom("TEST123")
			turnInfo := room.GetCurrentTurnInfo()

			if turnInfo.ClientID != "" {
				t.Errorf("Expected empty PeerInfo for no turn, got ClientID '%s'", turnInfo.ClientID)
			}
		})

		t.Run("ActiveTurn", func(t *testing.T) {
			room := NewRoom("TEST123")
			client := createTestClient("client1", "Alice", "#FF0000")
			client.TotalTurnTime = 1000
			room.AddClient(client)

			// Set turn using internal field (since SetCurrentTurn requires expectedCurrentTurn)
			room.mu.Lock()
			room.CurrentTurn = "client1"
			now := time.Now().UnixNano()
			room.TurnStartTime = &now
			room.mu.Unlock()

			turnInfo := room.GetCurrentTurnInfo()

			if turnInfo.ClientID != "client1" {
				t.Errorf("Expected ClientID 'client1', got '%s'", turnInfo.ClientID)
			}
			if turnInfo.DisplayName != "Alice" {
				t.Errorf("Expected DisplayName 'Alice', got '%s'", turnInfo.DisplayName)
			}
			if turnInfo.Color != "#FF0000" {
				t.Errorf("Expected Color '#FF0000', got '%s'", turnInfo.Color)
			}
			if turnInfo.TotalTurnTime != 1000 {
				t.Errorf("Expected TotalTurnTime 1000, got %d", turnInfo.TotalTurnTime)
			}
		})

		t.Run("TurnButClientNotFound", func(t *testing.T) {
			room := NewRoom("TEST123")
			// Set turn to non-existent client
			room.mu.Lock()
			room.CurrentTurn = "nonexistent"
			room.mu.Unlock()

			turnInfo := room.GetCurrentTurnInfo()

			if turnInfo.ClientID != "" {
				t.Errorf("Expected empty PeerInfo for missing client, got ClientID '%s'", turnInfo.ClientID)
			}
		})
	})

	t.Run("GetCurrentTurn", func(t *testing.T) {
		room := NewRoom("TEST123")

		// Initially no turn
		turn := room.GetCurrentTurn()
		if turn != "" {
			t.Errorf("Expected empty string for no turn, got '%s'", turn)
		}

		// Set turn
		room.mu.Lock()
		room.CurrentTurn = "client1"
		room.mu.Unlock()

		turn = room.GetCurrentTurn()
		if turn != "client1" {
			t.Errorf("Expected 'client1', got '%s'", turn)
		}
	})

	t.Run("GetTurnStartTime", func(t *testing.T) {
		t.Run("NoTurn", func(t *testing.T) {
			room := NewRoom("TEST123")
			startTime := room.GetTurnStartTime()
			if startTime != 0 {
				t.Errorf("Expected 0 for no turn, got %d", startTime)
			}
		})

		t.Run("WithTurn", func(t *testing.T) {
			room := NewRoom("TEST123")
			now := time.Now()
			nowNano := now.UnixNano()
			nowMs := now.UnixMilli()

			room.mu.Lock()
			room.TurnStartTime = &nowNano
			room.mu.Unlock()

			startTime := room.GetTurnStartTime()
			// Allow small difference due to timing (within 1 second)
			if startTime < nowMs-1000 || startTime > nowMs+1000 {
				t.Errorf("Expected TurnStartTime around %d ms, got %d ms (diff: %d)", nowMs, startTime, startTime-nowMs)
			}
		})
	})

	t.Run("SetCurrentTurn", func(t *testing.T) {
		t.Run("Success", func(t *testing.T) {
			room := NewRoom("TEST123")
			client1 := createTestClient("client1", "Alice", "#FF0000")
			client2 := createTestClient("client2", "Bob", "#00FF00")
			room.AddClient(client1)
			room.AddClient(client2)

			// Set initial turn
			if !room.SetCurrentTurn("", "client1") {
				t.Error("Expected SetCurrentTurn to succeed for first turn")
			}

			turn := room.GetCurrentTurn()
			if turn != "client1" {
				t.Errorf("Expected CurrentTurn 'client1', got '%s'", turn)
			}

			startTime1 := room.GetTurnStartTime()
			if startTime1 == 0 {
				t.Error("Expected TurnStartTime to be set")
			}

			// Wait a bit to ensure time difference
			time.Sleep(10 * time.Millisecond)

			// Change turn
			if !room.SetCurrentTurn("client1", "client2") {
				t.Error("Expected SetCurrentTurn to succeed for turn change")
			}

			turn = room.GetCurrentTurn()
			if turn != "client2" {
				t.Errorf("Expected CurrentTurn 'client2', got '%s'", turn)
			}

			startTime2 := room.GetTurnStartTime()
			if startTime2 <= startTime1 {
				t.Errorf("Expected new TurnStartTime (%d) > old (%d)", startTime2, startTime1)
			}

			// Verify client1's TotalTurnTime increased
			if client1.TotalTurnTime == 0 {
				t.Error("Expected client1's TotalTurnTime to be updated after turn change")
			}
		})

		t.Run("ClientNotFound", func(t *testing.T) {
			room := NewRoom("TEST123")
			if room.SetCurrentTurn("", "nonexistent") {
				t.Error("Expected SetCurrentTurn to fail for non-existent client")
			}
		})

		t.Run("StateMismatch", func(t *testing.T) {
			room := NewRoom("TEST123")
			client1 := createTestClient("client1", "Alice", "#FF0000")
			client2 := createTestClient("client2", "Bob", "#00FF00")
			room.AddClient(client1)
			room.AddClient(client2)

			// Set initial turn
			room.SetCurrentTurn("", "client1")

			// Try to set turn with wrong expectedCurrentTurn
			if room.SetCurrentTurn("wrong", "client2") {
				t.Error("Expected SetCurrentTurn to fail for state mismatch")
			}

			// Verify turn didn't change
			turn := room.GetCurrentTurn()
			if turn != "client1" {
				t.Errorf("Expected CurrentTurn to remain 'client1', got '%s'", turn)
			}
		})
	})

	t.Run("ClearCurrentTurn", func(t *testing.T) {
		t.Run("WithActiveTurn", func(t *testing.T) {
			room := NewRoom("TEST123")
			client := createTestClient("client1", "Alice", "#FF0000")
			room.AddClient(client)

			// Set turn
			room.SetCurrentTurn("", "client1")

			initialTotalTime := client.TotalTurnTime
			time.Sleep(10 * time.Millisecond)

			// Clear turn
			room.ClearCurrentTurn()

			turn := room.GetCurrentTurn()
			if turn != "" {
				t.Errorf("Expected empty CurrentTurn, got '%s'", turn)
			}

			startTime := room.GetTurnStartTime()
			if startTime != 0 {
				t.Errorf("Expected TurnStartTime 0, got %d", startTime)
			}

			// Verify client's TotalTurnTime increased
			if client.TotalTurnTime <= initialTotalTime {
				t.Errorf("Expected TotalTurnTime to increase, got %d (was %d)", client.TotalTurnTime, initialTotalTime)
			}
		})

		t.Run("WithoutTurn", func(t *testing.T) {
			room := NewRoom("TEST123")
			// Should not panic
			room.ClearCurrentTurn()

			turn := room.GetCurrentTurn()
			if turn != "" {
				t.Errorf("Expected empty CurrentTurn, got '%s'", turn)
			}
		})
	})

	t.Run("RemoveClient", func(t *testing.T) {
		t.Run("BasicRemoval", func(t *testing.T) {
			room := NewRoom("TEST123")
			client1 := createTestClient("client1", "Alice", "#FF0000")
			client2 := createTestClient("client2", "Bob", "#00FF00")
			room.AddClient(client1)
			room.AddClient(client2)

			hadTurn, isEmpty := room.RemoveClient("client1")

			if hadTurn {
				t.Error("Expected hadTurn false for client without turn")
			}
			if isEmpty {
				t.Error("Expected isEmpty false, room should still have client2")
			}

			if len(room.Clients) != 1 {
				t.Errorf("Expected 1 client remaining, got %d", len(room.Clients))
			}

			if room.Clients["client1"] != nil {
				t.Error("Expected client1 to be removed")
			}

			if room.Clients["client2"] != client2 {
				t.Error("Expected client2 to remain")
			}
		})

		t.Run("RemoveClientWithTurn", func(t *testing.T) {
			room := NewRoom("TEST123")
			client1 := createTestClient("client1", "Alice", "#FF0000")
			client2 := createTestClient("client2", "Bob", "#00FF00")
			room.AddClient(client1)
			room.AddClient(client2)

			// Set client1's turn
			room.SetCurrentTurn("", "client1")
			initialTotalTime := client1.TotalTurnTime
			time.Sleep(10 * time.Millisecond)

			hadTurn, isEmpty := room.RemoveClient("client1")

			if !hadTurn {
				t.Error("Expected hadTurn true for client with turn")
			}
			if isEmpty {
				t.Error("Expected isEmpty false, room should still have client2")
			}

			// Verify turn was cleared
			turn := room.GetCurrentTurn()
			if turn != "" {
				t.Errorf("Expected CurrentTurn to be cleared, got '%s'", turn)
			}

			// Verify client1's TotalTurnTime was updated
			if client1.TotalTurnTime <= initialTotalTime {
				t.Errorf("Expected TotalTurnTime to increase, got %d (was %d)", client1.TotalTurnTime, initialTotalTime)
			}
		})

		t.Run("RemoveLastClient", func(t *testing.T) {
			room := NewRoom("TEST123")
			client := createTestClient("client1", "Alice", "#FF0000")
			room.AddClient(client)

			hadTurn, isEmpty := room.RemoveClient("client1")

			if hadTurn {
				t.Error("Expected hadTurn false for client without turn")
			}
			if !isEmpty {
				t.Error("Expected isEmpty true for last client")
			}

			if len(room.Clients) != 0 {
				t.Errorf("Expected empty room, got %d clients", len(room.Clients))
			}
		})

		t.Run("RemoveNonExistentClient", func(t *testing.T) {
			room := NewRoom("TEST123")
			client := createTestClient("client1", "Alice", "#FF0000")
			room.AddClient(client)

			hadTurn, isEmpty := room.RemoveClient("nonexistent")

			if hadTurn {
				t.Error("Expected hadTurn false for non-existent client")
			}
			if isEmpty {
				t.Error("Expected isEmpty false, room should still have client1")
			}

			if len(room.Clients) != 1 {
				t.Errorf("Expected 1 client remaining, got %d", len(room.Clients))
			}
		})
	})

	t.Run("Concurrency", func(t *testing.T) {
		// Test that Room methods are thread-safe
		room := NewRoom("TEST123")
		client1 := createTestClient("client1", "Alice", "#FF0000")
		client2 := createTestClient("client2", "Bob", "#00FF00")
		room.AddClient(client1)
		room.AddClient(client2)

		// Run concurrent operations
		done := make(chan bool, 10)

		// Reader goroutines - only read operations
		for i := 0; i < 5; i++ {
			go func() {
				for j := 0; j < 100; j++ {
					room.GetCurrentTurn()
					room.ListPeerIDs()
					room.ListPeerInfo()
					room.GetTurnStartTime()
				}
				done <- true
			}()
		}

		// Writer goroutines - try to set/clear turns
		// State mismatches are expected and acceptable in concurrent scenarios
		for i := 0; i < 5; i++ {
			go func(id int) {
				for j := 0; j < 100; j++ {
					// Try to set turn - may fail due to concurrent modifications, that's ok
					currentTurn := room.GetCurrentTurn()
					// Only try to set if no turn is active, to reduce conflicts
					if currentTurn == "" {
						room.SetCurrentTurn("", "client1")
					} else {
						room.SetCurrentTurn(currentTurn, "client2")
					}
					// Clear turn - may fail if another goroutine changed it, that's ok
					room.ClearCurrentTurn()
				}
				done <- true
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 10; i++ {
			<-done
		}

		// Verify room is still in a valid state
		peers := room.ListPeerInfo()
		if len(peers) != 2 {
			t.Errorf("Expected 2 peers after concurrent operations, got %d", len(peers))
		}

		// Verify clients are still in room
		if room.Clients["client1"] != client1 {
			t.Error("Expected client1 to remain in room after concurrent operations")
		}
		if room.Clients["client2"] != client2 {
			t.Error("Expected client2 to remain in room after concurrent operations")
		}
	})
}
