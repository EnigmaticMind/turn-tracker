package updateprofile

import (
	"encoding/json"
	"testing"
	"time"

	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/createroom"
	"turn-tracker/backend/handlers/joinroom"
	"turn-tracker/backend/test_helpers"
	"turn-tracker/backend/types"
)

func setupTestMessageRouter() core.MessageHandler {
	return func(hub *core.Hub, client *core.Client, msg *types.Message) {
		switch msg.Type {
		case "create_room":
			var data createroom.CreateRoomData
			json.Unmarshal(msg.Data, &data)
			createroom.HandleCreateRoom(hub, client, data.RoomID, data.DisplayName, data.Color)
		case "join_room":
			var data joinroom.JoinRoomData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				errorMsg, _ := types.NewErrorMessage("Invalid join_room data")
				client.Send <- errorMsg
				return
			}
			joinroom.HandleJoinRoom(hub, client, data.RoomID, data.DisplayName, data.Color)
		case "update_profile":
			var data UpdateProfileData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				errorMsg, _ := types.NewErrorMessage("Invalid update_profile data")
				client.Send <- errorMsg
				return
			}
			HandleUpdateProfile(hub, client, data.DisplayName, data.Color)
		default:
			errorMsg, _ := types.NewUnknownMessageTypeError(msg.Type)
			client.Send <- errorMsg
		}
	}
}

// TestUpdateProfile wraps all update_profile tests
// This allows running all tests together or individually in the IDE
func TestUpdateProfile(t *testing.T) {
	t.Run("UpdatesDisplayName", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		time.Sleep(100 * time.Millisecond)

		// Create room
		err = client1.SendMessage("create_room", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		_, err = client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		// Update display name
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"display_name": "NewName",
		})
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		// Should receive broadcast (includes sender)
		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive broadcast: %v", err)
		}

		if resp.Type != "profile_updated" {
			t.Errorf("Expected message type 'profile_updated', got '%s'", resp.Type)
		}

		var data ProfileUpdatedData
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("Failed to unmarshal profile_updated: %v", err)
		}

		if data.DisplayName != "NewName" {
			t.Errorf("Expected DisplayName 'NewName', got '%s'", data.DisplayName)
		}
	})

	t.Run("UpdatesColor", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		time.Sleep(100 * time.Millisecond)

		// Create room
		err = client1.SendMessage("create_room", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		_, err = client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		// Update color
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"color": "#FF5733",
		})
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		// Should receive broadcast (includes sender)
		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive broadcast: %v", err)
		}

		var data ProfileUpdatedData
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("Failed to unmarshal profile_updated: %v", err)
		}

		if data.Color != "#FF5733" {
			t.Errorf("Expected Color '#FF5733', got '%s'", data.Color)
		}
	})

	t.Run("UpdatesBothDisplayNameAndColor", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		time.Sleep(100 * time.Millisecond)

		// Create room
		err = client1.SendMessage("create_room", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		_, err = client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		// Update both
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"display_name": "NewName",
			"color":        "#33FF57",
		})
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		// Should receive broadcast (includes sender)
		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive broadcast: %v", err)
		}

		var data ProfileUpdatedData
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("Failed to unmarshal profile_updated: %v", err)
		}

		if data.DisplayName != "NewName" {
			t.Errorf("Expected DisplayName 'NewName', got '%s'", data.DisplayName)
		}
		if data.Color != "#33FF57" {
			t.Errorf("Expected Color '#33FF57', got '%s'", data.Color)
		}
	})

	t.Run("BroadcastsToOtherPlayers", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		client2, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client2: %v", err)
		}
		defer client2.Close()

		time.Sleep(100 * time.Millisecond)

		// Client1 creates room
		err = client1.SendMessage("create_room", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		resp1, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		var roomData struct {
			RoomID string `json:"room_id"`
		}
		if err := json.Unmarshal(resp1.Data, &roomData); err != nil {
			t.Fatalf("Failed to unmarshal room_created: %v", err)
		}

		// Client2 joins room
		err = client2.SendMessage("join_room", map[string]interface{}{
			"room_id": roomData.RoomID,
		})
		if err != nil {
			t.Fatalf("Failed to join room: %v", err)
		}

		// Clear join messages
		_, _ = client2.ReceiveMessage(2 * time.Second)
		_, _ = client1.ReceiveMessage(2 * time.Second)

		// Client1 updates profile
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"display_name": "UpdatedName",
		})
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		// Both clients should receive broadcast
		broadcast1, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Client1 failed to receive broadcast: %v", err)
		}

		broadcast2, err := client2.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Client2 failed to receive broadcast: %v", err)
		}

		if broadcast1.Type != "profile_updated" {
			t.Errorf("Expected message type 'profile_updated' for client1, got '%s'", broadcast1.Type)
		}
		if broadcast2.Type != "profile_updated" {
			t.Errorf("Expected message type 'profile_updated' for client2, got '%s'", broadcast2.Type)
		}

		var data1, data2 ProfileUpdatedData
		if err := json.Unmarshal(broadcast1.Data, &data1); err != nil {
			t.Fatalf("Failed to unmarshal profile_updated (client1): %v", err)
		}
		if err := json.Unmarshal(broadcast2.Data, &data2); err != nil {
			t.Fatalf("Failed to unmarshal profile_updated (client2): %v", err)
		}

		if data1.DisplayName != "UpdatedName" {
			t.Errorf("Expected DisplayName 'UpdatedName' for client1, got '%s'", data1.DisplayName)
		}
		if data2.DisplayName != "UpdatedName" {
			t.Errorf("Expected DisplayName 'UpdatedName' for client2, got '%s'", data2.DisplayName)
		}
	})

	t.Run("ReturnsErrorWhenNotInRoom", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		time.Sleep(100 * time.Millisecond)

		// Try to update profile without being in a room
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"display_name": "NewName",
		})
		if err != nil {
			t.Fatalf("Failed to send update_profile: %v", err)
		}

		// Should receive error
		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive error: %v", err)
		}

		if resp.Type != "error" {
			t.Errorf("Expected message type 'error', got '%s'", resp.Type)
		}

		var errorData struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(resp.Data, &errorData); err != nil {
			t.Fatalf("Failed to unmarshal error: %v", err)
		}

		if errorData.Message != "Not in a room" {
			t.Errorf("Expected error message 'Not in a room', got '%s'", errorData.Message)
		}
	})

	t.Run("ReturnsErrorForInvalidColor", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		time.Sleep(100 * time.Millisecond)

		// Create room
		err = client1.SendMessage("create_room", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		_, err = client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		// Try to update with invalid color
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"color": "INVALID",
		})
		if err != nil {
			t.Fatalf("Failed to send update_profile: %v", err)
		}

		// Should receive error
		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive error: %v", err)
		}

		if resp.Type != "error" {
			t.Errorf("Expected message type 'error', got '%s'", resp.Type)
		}
	})

	t.Run("ReturnsErrorForEmptyDisplayName", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		time.Sleep(100 * time.Millisecond)

		// Create room
		err = client1.SendMessage("create_room", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		_, err = client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		// Try to update with empty display name (whitespace only)
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"display_name": "   ",
		})
		if err != nil {
			t.Fatalf("Failed to send update_profile: %v", err)
		}

		// Should receive error
		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive error: %v", err)
		}

		if resp.Type != "error" {
			t.Errorf("Expected message type 'error', got '%s'", resp.Type)
		}
	})

	t.Run("ReturnsErrorForDisplayNameTooLong", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		time.Sleep(100 * time.Millisecond)

		// Create room
		err = client1.SendMessage("create_room", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		_, err = client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		// Try to update with display name too long
		longName := ""
		for i := 0; i < 51; i++ {
			longName += "a"
		}
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"display_name": longName,
		})
		if err != nil {
			t.Fatalf("Failed to send update_profile: %v", err)
		}

		// Should receive error
		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive error: %v", err)
		}

		if resp.Type != "error" {
			t.Errorf("Expected message type 'error', got '%s'", resp.Type)
		}
	})

	t.Run("ReturnsEarlyWhenNoUpdatesNeeded", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		time.Sleep(100 * time.Millisecond)

		// Create room
		err = client1.SendMessage("create_room", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		var roomData struct {
			Peers []struct {
				ClientID    string `json:"client_id"`
				DisplayName string `json:"display_name"`
				Color       string `json:"color"`
			} `json:"peers"`
		}
		if err := json.Unmarshal(resp.Data, &roomData); err != nil {
			t.Fatalf("Failed to unmarshal room_created: %v", err)
		}

		if len(roomData.Peers) == 0 {
			t.Fatal("Expected at least one peer")
		}
		currentName := roomData.Peers[0].DisplayName
		currentColor := roomData.Peers[0].Color

		// Send update with same values (should return early, no message)
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"display_name": currentName,
			"color":        currentColor,
		})
		if err != nil {
			t.Fatalf("Failed to send update_profile: %v", err)
		}

		// Should NOT receive any message (early return)
		resp, err = client1.ReceiveMessage(500 * time.Millisecond)
		if err == nil {
			t.Errorf("Expected no message when no update needed, but received: %s", resp.Type)
		}
	})

	t.Run("NormalizesColorToUpperCase", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client1, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect client1: %v", err)
		}
		defer client1.Close()

		time.Sleep(100 * time.Millisecond)

		// Create room
		err = client1.SendMessage("create_room", map[string]interface{}{})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		_, err = client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		// Update with lowercase color
		err = client1.SendMessage("update_profile", map[string]interface{}{
			"color": "#ff5733",
		})
		if err != nil {
			t.Fatalf("Failed to update profile: %v", err)
		}

		// Should receive broadcast with uppercase color
		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive broadcast: %v", err)
		}

		var data ProfileUpdatedData
		if err := json.Unmarshal(resp.Data, &data); err != nil {
			t.Fatalf("Failed to unmarshal profile_updated: %v", err)
		}

		if data.Color != "#FF5733" {
			t.Errorf("Expected Color '#FF5733', got '%s'", data.Color)
		}
	})
}
