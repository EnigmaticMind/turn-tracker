package core

import (
	"sync/atomic"
	"testing"
	"time"
)

// TestUnregister wraps all unregister tests
// This allows running all tests together or individually in the IDE
func TestUnregister(t *testing.T) {
	// Helper to create a test client (Conn is nil - unregister.go handles it safely)
	createTestClient := func(clientID, roomID, displayName, color string) *Client {
		return &Client{
			ClientID:    clientID,
			RoomID:      roomID,
			DisplayName: displayName,
			Color:       color,
			Send:        make(chan []byte, 32),
			Conn:        nil, // nil is safe - unregister.go uses "<no-connection>" when nil
		}
	}

	t.Run("RemovesClientFromHub", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("test-client-1", "", "", "")

		hub.clients[client] = true
		if len(hub.clients) != 1 {
			t.Fatal("Expected 1 client before unregister")
		}

		hub.handleUnregister(client)

		if len(hub.clients) != 0 {
			t.Error("Expected client to be removed from hub.clients")
		}
	})

	t.Run("EarlyReturnIfClientNotRegistered", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("test-client-2", "", "", "")

		// Should not panic or error
		hub.handleUnregister(client)

		// Connection counter should not change
		if atomic.LoadInt32(&hub.currentConnections) != 0 {
			t.Error("Expected currentConnections to remain 0 for unregistered client")
		}
	})

	t.Run("DecrementsConnectionCounter", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("test-client-3", "", "", "")

		// Increment counter first
		hub.TryRegister()
		if atomic.LoadInt32(&hub.currentConnections) != 1 {
			t.Fatal("Expected currentConnections to be 1")
		}

		hub.clients[client] = true
		hub.handleUnregister(client)

		if atomic.LoadInt32(&hub.currentConnections) != 0 {
			t.Error("Expected currentConnections to be decremented to 0")
		}
	})

	t.Run("ClosesClientSendChannel", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("test-client-4", "", "", "")

		hub.clients[client] = true
		hub.handleUnregister(client)

		// Verify channel is actually closed
		if _, ok := <-client.Send; ok {
			t.Error("Expected Send channel to be closed (receive should return false)")
		}
	})

	t.Run("SavesDisconnectedClientDataWhenInRoomWithProfile", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("test-client-5", "ROOM123", "Alice", "#FF0000")

		hub.clients[client] = true
		hub.handleUnregister(client)

		hub.disconnectedMu.RLock()
		disconnected, exists := hub.disconnectedClients["test-client-5"]
		hub.disconnectedMu.RUnlock()

		if !exists {
			t.Error("Expected disconnected client data to be saved when in room with profile")
		}
		if disconnected == nil {
			t.Fatal("Disconnected client should not be nil")
		}
		if disconnected.ClientID != "test-client-5" {
			t.Errorf("Expected ClientID 'test-client-5', got '%s'", disconnected.ClientID)
		}
		if disconnected.DisplayName != "Alice" {
			t.Errorf("Expected DisplayName 'Alice', got '%s'", disconnected.DisplayName)
		}
		if disconnected.Color != "#FF0000" {
			t.Errorf("Expected Color '#FF0000', got '%s'", disconnected.Color)
		}
		if disconnected.LastRoomID != "ROOM123" {
			t.Errorf("Expected LastRoomID 'ROOM123', got '%s'", disconnected.LastRoomID)
		}
	})

	t.Run("DoesNotSaveDataWhenNoRoomID", func(t *testing.T) {
		hub := NewHub()
		// No roomID = client never joined a room
		client := createTestClient("test-client-6", "", "Alice", "#FF0000")

		hub.clients[client] = true
		hub.handleUnregister(client)

		hub.disconnectedMu.RLock()
		_, exists := hub.disconnectedClients["test-client-6"]
		hub.disconnectedMu.RUnlock()

		if exists {
			t.Error("Expected disconnected client data NOT to be saved when no roomID")
		}
	})

	t.Run("SkipsSavingIfNoClientID", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("", "ROOM123", "Alice", "#FF0000")

		hub.clients[client] = true
		hub.handleUnregister(client)

		hub.disconnectedMu.RLock()
		count := len(hub.disconnectedClients)
		hub.disconnectedMu.RUnlock()

		if count != 0 {
			t.Error("Expected disconnected client data NOT to be saved (no clientID)")
		}
	})

	t.Run("SkipsSavingIfNoDisplayNameOrColor", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("test-client-7", "ROOM123", "", "")

		hub.clients[client] = true
		hub.handleUnregister(client)

		hub.disconnectedMu.RLock()
		_, exists := hub.disconnectedClients["test-client-7"]
		hub.disconnectedMu.RUnlock()

		if exists {
			t.Error("Expected disconnected client data NOT to be saved (no display name or color)")
		}
	})

	t.Run("EarlyReturnIfNoRoomID", func(t *testing.T) {
		hub := NewHub()
		room := NewRoom("ROOM123")
		client := createTestClient("client-1", "", "Alice", "#FF0000") // No roomID

		room.Clients[client.ClientID] = client

		hub.mu.Lock()
		hub.rooms["ROOM123"] = room
		hub.mu.Unlock()

		hub.clients[client] = true

		// These should not be called when roomID is empty
		turnEndedCalled := false
		playerLeftCalled := false
		hub.OnTurnEnded = func(roomID string) {
			turnEndedCalled = true
		}
		hub.OnPlayerLeft = func(roomID, clientID string, message []byte) {
			playerLeftCalled = true
		}

		hub.handleUnregister(client)

		if turnEndedCalled {
			t.Error("Expected OnTurnEnded NOT to be called when roomID is empty")
		}
		if playerLeftCalled {
			t.Error("Expected OnPlayerLeft NOT to be called when roomID is empty")
		}

		// Room should still exist (client wasn't removed from it)
		hub.mu.RLock()
		_, exists := hub.rooms["ROOM123"]
		hub.mu.RUnlock()
		if !exists {
			t.Error("Expected room to still exist (client not removed because no roomID)")
		}
	})

	t.Run("RemovesClientFromRoom", func(t *testing.T) {
		hub := NewHub()
		room := NewRoom("ROOM123")
		client1 := createTestClient("client-1", "ROOM123", "Alice", "#FF0000")
		client2 := createTestClient("client-2", "ROOM123", "Bob", "#00FF00")

		room.Clients[client1.ClientID] = client1
		room.Clients[client2.ClientID] = client2

		hub.mu.Lock()
		hub.rooms["ROOM123"] = room
		hub.mu.Unlock()

		hub.clients[client1] = true
		hub.handleUnregister(client1)

		room.mu.RLock()
		_, exists := room.Clients[client1.ClientID]
		client2Exists := room.Clients[client2.ClientID] != nil
		room.mu.RUnlock()

		if exists {
			t.Error("Expected client1 to be removed from room")
		}
		if !client2Exists {
			t.Error("Expected client2 to remain in room")
		}
	})

	t.Run("DoesNotDeleteEmptyRoom", func(t *testing.T) {
		hub := NewHub()
		room := NewRoom("ROOM123")
		client := createTestClient("client-1", "ROOM123", "Alice", "#FF0000")

		room.Clients[client.ClientID] = client

		hub.mu.Lock()
		hub.rooms["ROOM123"] = room
		hub.mu.Unlock()

		hub.clients[client] = true
		hub.handleUnregister(client)

		hub.mu.RLock()
		_, exists := hub.rooms["ROOM123"]
		hub.mu.RUnlock()

		if !exists {
			t.Error("Expected empty room to NOT be deleted (scheduled cleanup will handle it)")
		}
	})

	t.Run("DoesNotCallCallbacksWhenRoomBecomesEmpty", func(t *testing.T) {
		hub := NewHub()
		room := NewRoom("ROOM123")
		client := createTestClient("client-1", "ROOM123", "Alice", "#FF0000")

		room.Clients[client.ClientID] = client
		room.CurrentTurn = "client-1"
		now := time.Now().UnixNano()
		room.TurnStartTime = &now

		hub.mu.Lock()
		hub.rooms["ROOM123"] = room
		hub.mu.Unlock()

		hub.clients[client] = true

		turnEndedCalled := false
		playerLeftCalled := false
		hub.OnTurnEnded = func(roomID string) {
			turnEndedCalled = true
		}
		hub.OnPlayerLeft = func(roomID, clientID string, message []byte) {
			playerLeftCalled = true
		}

		hub.handleUnregister(client)

		// Room is now empty, so callbacks should NOT be called
		if turnEndedCalled {
			t.Error("Expected OnTurnEnded NOT to be called when room becomes empty")
		}
		if playerLeftCalled {
			t.Error("Expected OnPlayerLeft NOT to be called when room becomes empty")
		}
	})

	t.Run("CallsOnTurnEndedWhenClientHadTurnAndRoomNotEmpty", func(t *testing.T) {
		hub := NewHub()
		room := NewRoom("ROOM123")
		client1 := createTestClient("client-1", "ROOM123", "Alice", "#FF0000")
		client2 := createTestClient("client-2", "ROOM123", "Bob", "#00FF00")

		room.Clients[client1.ClientID] = client1
		room.Clients[client2.ClientID] = client2
		room.CurrentTurn = "client-1"
		now := time.Now().UnixNano()
		room.TurnStartTime = &now

		hub.mu.Lock()
		hub.rooms["ROOM123"] = room
		hub.mu.Unlock()

		hub.clients[client1] = true

		turnEndedCalled := false
		var calledRoomID string
		hub.OnTurnEnded = func(roomID string) {
			turnEndedCalled = true
			calledRoomID = roomID
		}

		hub.handleUnregister(client1)

		if !turnEndedCalled {
			t.Error("Expected OnTurnEnded callback to be called when client had turn and room is not empty")
		}
		if calledRoomID != "ROOM123" {
			t.Errorf("Expected OnTurnEnded to be called with 'ROOM123', got '%s'", calledRoomID)
		}
	})

	t.Run("CallsOnPlayerLeftWhenRoomNotEmpty", func(t *testing.T) {
		hub := NewHub()
		room := NewRoom("ROOM123")
		client1 := createTestClient("client-1", "ROOM123", "Alice", "#FF0000")
		client2 := createTestClient("client-2", "ROOM123", "Bob", "#00FF00")

		room.Clients[client1.ClientID] = client1
		room.Clients[client2.ClientID] = client2

		hub.mu.Lock()
		hub.rooms["ROOM123"] = room
		hub.mu.Unlock()

		hub.clients[client1] = true

		playerLeftCalled := false
		var calledRoomID, calledClientID string
		hub.OnPlayerLeft = func(roomID, clientID string, message []byte) {
			playerLeftCalled = true
			calledRoomID = roomID
			calledClientID = clientID
		}

		hub.handleUnregister(client1)

		if !playerLeftCalled {
			t.Error("Expected OnPlayerLeft callback to be called when room is not empty")
		}
		if calledRoomID != "ROOM123" {
			t.Errorf("Expected OnPlayerLeft to be called with roomID 'ROOM123', got '%s'", calledRoomID)
		}
		if calledClientID != "client-1" {
			t.Errorf("Expected OnPlayerLeft to be called with clientID 'client-1', got '%s'", calledClientID)
		}
	})

	t.Run("DoesNotCallOnTurnEndedWhenClientDidNotHaveTurn", func(t *testing.T) {
		hub := NewHub()
		room := NewRoom("ROOM123")
		client1 := createTestClient("client-1", "ROOM123", "Alice", "#FF0000")
		client2 := createTestClient("client-2", "ROOM123", "Bob", "#00FF00")

		room.Clients[client1.ClientID] = client1
		room.Clients[client2.ClientID] = client2
		room.CurrentTurn = "client-2" // client1 does not have turn

		hub.mu.Lock()
		hub.rooms["ROOM123"] = room
		hub.mu.Unlock()

		hub.clients[client1] = true

		turnEndedCalled := false
		hub.OnTurnEnded = func(roomID string) {
			turnEndedCalled = true
		}

		hub.handleUnregister(client1)

		if turnEndedCalled {
			t.Error("Expected OnTurnEnded NOT to be called when client did not have turn")
		}
	})

	t.Run("HandlesNilRoomGracefully", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("client-1", "NONEXISTENT", "Alice", "#FF0000")

		hub.clients[client] = true

		// Should not panic
		hub.handleUnregister(client)

		// Client should still be removed
		if len(hub.clients) != 0 {
			t.Error("Expected client to be removed even if room doesn't exist")
		}
	})

	t.Run("HandlesNilCallbacksGracefully", func(t *testing.T) {
		hub := NewHub()
		room := NewRoom("ROOM123")
		client1 := createTestClient("client-1", "ROOM123", "Alice", "#FF0000")
		client2 := createTestClient("client-2", "ROOM123", "Bob", "#00FF00")

		room.Clients[client1.ClientID] = client1
		room.Clients[client2.ClientID] = client2

		hub.mu.Lock()
		hub.rooms["ROOM123"] = room
		hub.mu.Unlock()

		hub.clients[client1] = true
		hub.OnTurnEnded = nil
		hub.OnPlayerLeft = nil

		// Should not panic
		hub.handleUnregister(client1)
	})

	t.Run("PreservesTotalTurnTime", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("client-1", "ROOM123", "Alice", "#FF0000")
		client.TotalTurnTime = 5000 // 5 seconds in milliseconds

		hub.clients[client] = true
		hub.handleUnregister(client)

		hub.disconnectedMu.RLock()
		disconnected, exists := hub.disconnectedClients["client-1"]
		hub.disconnectedMu.RUnlock()

		if !exists {
			t.Fatal("Expected disconnected client data to be saved")
		}
		if disconnected.TotalTurnTime != 5000 {
			t.Errorf("Expected TotalTurnTime 5000, got %d", disconnected.TotalTurnTime)
		}
	})

	t.Run("RecordsDisconnectedAtTime", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("client-1", "ROOM123", "Alice", "#FF0000")

		before := time.Now()
		hub.clients[client] = true
		hub.handleUnregister(client)
		after := time.Now()

		hub.disconnectedMu.RLock()
		disconnected, exists := hub.disconnectedClients["client-1"]
		hub.disconnectedMu.RUnlock()

		if !exists {
			t.Fatal("Expected disconnected client data to be saved")
		}

		if disconnected.DisconnectedAt.Before(before) || disconnected.DisconnectedAt.After(after) {
			t.Errorf("Expected DisconnectedAt to be between %v and %v, got %v",
				before, after, disconnected.DisconnectedAt)
		}
	})
}
