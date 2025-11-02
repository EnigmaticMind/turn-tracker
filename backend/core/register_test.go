package core

import (
	"sync"
	"testing"
	"time" // Add this import
)

// TestRegister wraps all register tests
// This allows running all tests together or individually in the IDE
func TestRegister(t *testing.T) {
	// Helper to create a test client
	createTestClient := func(clientID string) *Client {
		return &Client{
			ClientID: clientID,
			Conn:     nil, // nil is safe - register.go handles it
			Send:     make(chan []byte, 32),
		}
	}

	t.Run("GeneratesClientIDWhenEmpty", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("")

		hub.handleRegister(client)

		if client.ClientID == "" {
			t.Error("Expected clientID to be generated when empty")
		}
		if len(client.ClientID) != 16 {
			t.Errorf("Expected clientID to be 16 characters, got %d", len(client.ClientID))
		}
		// Note: hub.clients is not protected by hub.mu (only accessed from hub.Run() goroutine)
		// In tests we access it directly, so no mutex needed
		if len(hub.clients) != 1 {
			t.Error("Expected client to be added to hub")
		}
	})

	t.Run("RestoresDataOnReconnection", func(t *testing.T) {
		hub := NewHub()
		clientID := "abc123def4567890"
		client := createTestClient(clientID)

		// Add to disconnected clients
		disconnected := &DisconnectedClient{
			ClientID:      clientID,
			DisplayName:   "TestUser",
			Color:         "#FF0000",
			TotalTurnTime: 5000,
		}

		hub.disconnectedMu.Lock()
		hub.disconnectedClients[clientID] = disconnected
		hub.disconnectedMu.Unlock()

		hub.handleRegister(client)

		if client.DisplayName != "TestUser" {
			t.Errorf("Expected DisplayName 'TestUser', got '%s'", client.DisplayName)
		}
		if client.Color != "#FF0000" {
			t.Errorf("Expected Color '#FF0000', got '%s'", client.Color)
		}
		if client.TotalTurnTime != 5000 {
			t.Errorf("Expected TotalTurnTime 5000, got %d", client.TotalTurnTime)
		}

		// Verify removed from disconnected clients
		hub.disconnectedMu.RLock()
		_, exists := hub.disconnectedClients[clientID]
		hub.disconnectedMu.RUnlock()
		if exists {
			t.Error("Expected client to be removed from disconnectedClients on reconnect")
		}
	})

	t.Run("GeneratesNewIDForInvalidFormat", func(t *testing.T) {
		hub := NewHub()
		invalidID := "invalid-id-format"
		client := createTestClient(invalidID)

		hub.handleRegister(client)

		if client.ClientID == invalidID {
			t.Error("Expected invalid clientID to be replaced")
		}
		if len(client.ClientID) != 16 {
			t.Errorf("Expected new clientID to be 16 characters, got %d", len(client.ClientID))
		}
	})

	t.Run("GeneratesNewIDForWrongLength", func(t *testing.T) {
		hub := NewHub()
		shortID := "abc123" // Too short
		client := createTestClient(shortID)

		hub.handleRegister(client)

		if client.ClientID == shortID {
			t.Error("Expected short clientID to be replaced")
		}
		if len(client.ClientID) != 16 {
			t.Errorf("Expected new clientID to be 16 characters, got %d", len(client.ClientID))
		}
	})

	t.Run("GeneratesNewIDForInvalidCharacters", func(t *testing.T) {
		hub := NewHub()
		invalidID := "abc123def456789g" // Has 'g' which is invalid hex
		client := createTestClient(invalidID)

		hub.handleRegister(client)

		if client.ClientID == invalidID {
			t.Error("Expected invalid clientID to be replaced")
		}
		if len(client.ClientID) != 16 {
			t.Errorf("Expected new clientID to be 16 characters, got %d", len(client.ClientID))
		}
	})

	t.Run("KeepsValidClientID", func(t *testing.T) {
		hub := NewHub()
		validID := "abc123def4567890" // 16 hex chars
		client := createTestClient(validID)

		hub.handleRegister(client)

		if client.ClientID != validID {
			t.Errorf("Expected valid clientID to be kept, got '%s'", client.ClientID)
		}
	})

	t.Run("KeepsValidClientIDNotInDisconnected", func(t *testing.T) {
		hub := NewHub()
		validID := "1234567890abcdef" // 16 hex chars, not in disconnected
		client := createTestClient(validID)

		hub.handleRegister(client)

		if client.ClientID != validID {
			t.Errorf("Expected valid clientID to be kept when not in disconnected, got '%s'", client.ClientID)
		}
	})

	t.Run("HandlesNilConnSafely", func(t *testing.T) {
		hub := NewHub()
		client := createTestClient("")

		// Should not panic
		hub.handleRegister(client)

		if len(hub.clients) != 1 {
			t.Error("Expected client to be registered despite nil Conn")
		}
	})

	t.Run("SequentialRegistrations", func(t *testing.T) {
		hub := NewHub()
		const numClients = 100

		for i := 0; i < numClients; i++ {
			client := createTestClient("")
			hub.handleRegister(client)
		}

		if len(hub.clients) != numClients {
			t.Errorf("Expected %d clients, got %d", numClients, len(hub.clients))
		}
	})

	t.Run("ConcurrentRegistrationsViaChannel", func(t *testing.T) {
		hub := NewHub()

		// Start hub.Run() in a goroutine
		go hub.Run()

		const numClients = 100
		var wg sync.WaitGroup
		wg.Add(numClients)

		// Send clients through the Register channel (realistic approach)
		for i := 0; i < numClients; i++ {
			go func() {
				defer wg.Done()
				client := createTestClient("")
				hub.Register <- client
			}()
		}

		wg.Wait()

		// Give hub.Run() time to process all registrations
		// Wait a bit for channel processing
		time.Sleep(100 * time.Millisecond)

		if len(hub.clients) != numClients {
			t.Errorf("Expected %d clients, got %d", numClients, len(hub.clients))
		}
	})

	t.Run("ConcurrentReconnections", func(t *testing.T) {
		hub := NewHub()
		clientID := "abc123def4567890"

		// Setup disconnected client
		disconnected := &DisconnectedClient{
			ClientID:    clientID,
			DisplayName: "User",
			Color:       "#FF0000",
		}
		hub.disconnectedMu.Lock()
		hub.disconnectedClients[clientID] = disconnected
		hub.disconnectedMu.Unlock()

		// Start hub.Run() in a goroutine
		go hub.Run()

		const numClients = 10
		var wg sync.WaitGroup
		wg.Add(numClients)

		// Send clients through Register channel
		for i := 0; i < numClients; i++ {
			go func() {
				defer wg.Done()
				client := createTestClient(clientID)
				hub.Register <- client
			}()
		}

		wg.Wait()

		// Give hub.Run() time to process
		time.Sleep(100 * time.Millisecond)

		if len(hub.clients) != numClients {
			t.Errorf("Expected %d clients, got %d", numClients, len(hub.clients))
		}
	})
}

// TestIsValidClientID tests the validation function
func TestIsValidClientID(t *testing.T) {
	tests := []struct {
		name     string
		clientID string
		want     bool
	}{
		{"Valid16Hex", "abc123def4567890", true},
		{"Valid16HexUppercase", "ABC123DEF4567890", true},
		{"Valid16HexMixed", "aBc123DeF4567890", true},
		{"Valid16HexNumbers", "0123456789abcdef", true},
		{"TooShort", "abc123def45678", false},
		{"TooLong", "abc123def45678901", false},
		{"InvalidCharG", "abc123def456789g", false},
		{"InvalidCharZ", "abc123def456789z", false},
		{"InvalidCharSpecial", "abc123def456789!", false},
		{"InvalidCharSpace", "abc123def45678 0", false},
		{"Empty", "", false},
		{"Short14Chars", "abc123def45678", false},
		{"Long17Chars", "abc123def45678901", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidClientID(tt.clientID); got != tt.want {
				t.Errorf("isValidClientID(%q) = %v, want %v", tt.clientID, got, tt.want)
			}
		})
	}
}
