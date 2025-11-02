package leaveroom

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/createroom"
	"turn-tracker/backend/handlers/joinroom"
	"turn-tracker/backend/handlers/startturn"
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
			roomID := strings.ToUpper(data.RoomID)
			joinroom.HandleJoinRoom(hub, client, roomID, data.DisplayName, data.Color)
		case "leave_room":
			var data LeaveRoomData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				errorMsg, _ := types.NewErrorMessage("Invalid leave_room data")
				client.Send <- errorMsg
				return
			}
			roomID := strings.ToUpper(data.RoomID)
			HandleLeaveRoom(hub, client, roomID)
		case "start_turn":
			var data startturn.StartTurnData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				errorMsg, _ := types.NewErrorMessage("Invalid start_turn data")
				client.Send <- errorMsg
				return
			}
			startturn.HandleStartTurn(hub, client, data.CurrentTurn, data.NewTurn)
		default:
			errorMsg, _ := types.NewUnknownMessageTypeError(msg.Type)
			client.Send <- errorMsg
		}
	}
}

// TestLeaveRoom wraps all leave_room tests
// This allows running all tests together or individually in the IDE
func TestLeaveRoom(t *testing.T) {
	t.Run("LeavesRoomSuccessfully", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		hub := server.Hub
		hub.OnPlayerLeft = func(roomID, clientID string, msg []byte) {}
		hub.OnTurnEnded = func(roomID string) {}

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
			RoomID string `json:"room_id"`
		}
		if err := json.Unmarshal(resp.Data, &roomData); err != nil {
			t.Fatalf("Failed to unmarshal room_created: %v", err)
		}

		// Leave room
		err = client1.SendMessage("leave_room", map[string]interface{}{
			"room_id": roomData.RoomID,
		})
		if err != nil {
			t.Fatalf("Failed to send leave_room: %v", err)
		}

		// No confirmation message is sent, just verify room is empty
		// Wait a bit for processing
		time.Sleep(100 * time.Millisecond)

		// Verify room is now empty
		room := hub.GetRoom(roomData.RoomID)
		if room != nil {
			if len(room.Clients) != 0 {
				t.Errorf("Expected room to be empty, but has %d clients", len(room.Clients))
			}
		}

		// Note: client.RoomID is NOT cleared by leave_room (per handler comment)
		// This allows handleUnregister to handle the disconnect normally
		// We can't verify this directly as hub.clients is unexported
	})

	t.Run("DoesNotCallCallbacksWhenRoomBecomesEmpty", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		playerLeftCalled := false
		turnEndedCalled := false

		hub := server.Hub
		hub.OnPlayerLeft = func(roomID, clientID string, msg []byte) {
			playerLeftCalled = true
		}
		hub.OnTurnEnded = func(roomID string) {
			turnEndedCalled = true
		}

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
			RoomID string `json:"room_id"`
		}
		if err := json.Unmarshal(resp.Data, &roomData); err != nil {
			t.Fatalf("Failed to unmarshal room_created: %v", err)
		}

		// Leave room (this makes room empty)
		err = client1.SendMessage("leave_room", map[string]interface{}{
			"room_id": roomData.RoomID,
		})
		if err != nil {
			t.Fatalf("Failed to send leave_room: %v", err)
		}

		// Wait a bit for processing
		time.Sleep(100 * time.Millisecond)

		// Callbacks should NOT be called when room becomes empty
		if playerLeftCalled {
			t.Error("Expected OnPlayerLeft callback NOT to be called when room becomes empty")
		}
		if turnEndedCalled {
			t.Error("Expected OnTurnEnded callback NOT to be called when room becomes empty")
		}
	})

	t.Run("NotifiesOtherPlayersWhenLeaving", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		playerLeftCalled := false
		var leftRoomID, leftClientID string

		hub := server.Hub
		hub.OnPlayerLeft = func(roomID, clientID string, msg []byte) {
			playerLeftCalled = true
			leftRoomID = roomID
			leftClientID = clientID
		}
		hub.OnTurnEnded = func(roomID string) {}

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

		// Client2 leaves room
		err = client2.SendMessage("leave_room", map[string]interface{}{
			"room_id": roomData.RoomID,
		})
		if err != nil {
			t.Fatalf("Failed to send leave_room: %v", err)
		}

		// Wait a bit for callback to fire
		time.Sleep(100 * time.Millisecond)

		if !playerLeftCalled {
			t.Error("Expected OnPlayerLeft callback to be called")
		}
		if leftRoomID != roomData.RoomID {
			t.Errorf("Expected roomID %s, got %s", roomData.RoomID, leftRoomID)
		}
		if leftClientID == "" {
			t.Error("Expected clientID to be set in callback")
		}
	})

	t.Run("CallsOnTurnEndedWhenLeavingWithCurrentTurn", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		turnEndedCalled := false
		var endedRoomID string

		hub := server.Hub
		hub.OnPlayerLeft = func(roomID, clientID string, msg []byte) {}
		hub.OnTurnEnded = func(roomID string) {
			turnEndedCalled = true
			endedRoomID = roomID
		}

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

		// Get client2ID from room_joined message (creator is last, so joiner is first)
		roomJoined, err := client2.ReceiveMessage(2 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_joined: %v", err)
		}

		var joinResp struct {
			Peers []struct {
				ClientID string `json:"client_id"`
			} `json:"peers"`
		}
		var client2ID string
		if err := json.Unmarshal(roomJoined.Data, &joinResp); err == nil && len(joinResp.Peers) > 0 {
			// Creator is last, so joiner (client2) is first
			client2ID = joinResp.Peers[0].ClientID
		} else {
			// Fallback: get from player_joined
			playerJoined, err := client1.ReceiveMessage(2 * time.Second)
			if err != nil {
				t.Fatalf("Failed to receive player_joined: %v", err)
			}
			var pj struct {
				PeerID string `json:"peer_id"`
			}
			if err := json.Unmarshal(playerJoined.Data, &pj); err == nil {
				client2ID = pj.PeerID
			}
		}

		if client2ID == "" {
			t.Fatal("Could not determine client2ID")
		}

		// Start turn for client2
		err = client1.SendMessage("start_turn", map[string]interface{}{
			"current_turn": "",
			"new_turn":     client2ID,
		})
		if err != nil {
			t.Fatalf("Failed to start turn: %v", err)
		}

		// Clear turn_changed messages
		_, err = client1.ReceiveMessage(2 * time.Second)
		if err != nil {
			t.Fatalf("Client1 failed to receive turn_changed: %v", err)
		}
		_, err = client2.ReceiveMessage(2 * time.Second)
		if err != nil {
			t.Fatalf("Client2 failed to receive turn_changed: %v", err)
		}

		// Client2 leaves room (while having current turn)
		err = client2.SendMessage("leave_room", map[string]interface{}{
			"room_id": roomData.RoomID,
		})
		if err != nil {
			t.Fatalf("Failed to send leave_room: %v", err)
		}

		// Wait a bit for callback to fire
		time.Sleep(100 * time.Millisecond)

		if !turnEndedCalled {
			t.Error("Expected OnTurnEnded callback to be called when player with turn leaves")
		}
		if endedRoomID != roomData.RoomID {
			t.Errorf("Expected roomID %s, got %s", roomData.RoomID, endedRoomID)
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

		// Try to leave room without being in one
		err = client1.SendMessage("leave_room", map[string]interface{}{
			"room_id": "TEST1",
		})
		if err != nil {
			t.Fatalf("Failed to send leave_room: %v", err)
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

	t.Run("ReturnsErrorForRoomIDMismatch", func(t *testing.T) {
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
			RoomID string `json:"room_id"`
		}
		if err := json.Unmarshal(resp.Data, &roomData); err != nil {
			t.Fatalf("Failed to unmarshal room_created: %v", err)
		}

		// Try to leave with wrong room ID
		err = client1.SendMessage("leave_room", map[string]interface{}{
			"room_id": "WRONG",
		})
		if err != nil {
			t.Fatalf("Failed to send leave_room: %v", err)
		}

		// Should receive error
		errorResp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive error: %v", err)
		}

		if errorResp.Type != "error" {
			t.Errorf("Expected message type 'error', got '%s'", errorResp.Type)
		}

		var errorData struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(errorResp.Data, &errorData); err != nil {
			t.Fatalf("Failed to unmarshal error: %v", err)
		}

		if errorData.Message != "Room ID mismatch" {
			t.Errorf("Expected error message 'Room ID mismatch', got '%s'", errorData.Message)
		}
	})

	t.Run("ReturnsErrorForRoomNotFound", func(t *testing.T) {
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
			RoomID string `json:"room_id"`
		}
		if err := json.Unmarshal(resp.Data, &roomData); err != nil {
			t.Fatalf("Failed to unmarshal room_created: %v", err)
		}

		// Join the room first to set client.RoomID
		err = client1.SendMessage("join_room", map[string]interface{}{
			"room_id": roomData.RoomID,
		})
		if err != nil {
			t.Fatalf("Failed to join room: %v", err)
		}

		// Clear join message
		_, _ = client1.ReceiveMessage(2 * time.Second)

		// Delete the room to simulate it not existing
		hub := server.Hub
		hub.DeleteRoom(roomData.RoomID)

		// Try to leave non-existent room
		err = client1.SendMessage("leave_room", map[string]interface{}{
			"room_id": roomData.RoomID,
		})
		if err != nil {
			t.Fatalf("Failed to send leave_room: %v", err)
		}

		// Should receive error
		errorResp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive error: %v", err)
		}

		if errorResp.Type != "error" {
			t.Errorf("Expected message type 'error', got '%s'", errorResp.Type)
		}

		var errorData struct {
			Message string `json:"message"`
		}
		if err := json.Unmarshal(errorResp.Data, &errorData); err != nil {
			t.Fatalf("Failed to unmarshal error: %v", err)
		}

		if errorData.Message != "Room not found" {
			t.Errorf("Expected error message 'Room not found', got '%s'", errorData.Message)
		}
	})

	t.Run("NormalizesRoomIDToUppercase", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		hub := server.Hub
		hub.OnPlayerLeft = func(roomID, clientID string, msg []byte) {}
		hub.OnTurnEnded = func(roomID string) {}

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
			RoomID string `json:"room_id"`
		}
		if err := json.Unmarshal(resp.Data, &roomData); err != nil {
			t.Fatalf("Failed to unmarshal room_created: %v", err)
		}

		// Leave room with lowercase roomID
		err = client1.SendMessage("leave_room", map[string]interface{}{
			"room_id": strings.ToLower(roomData.RoomID),
		})
		if err != nil {
			t.Fatalf("Failed to send leave_room: %v", err)
		}

		// Wait a bit for processing
		time.Sleep(100 * time.Millisecond)

		// Verify room is empty (normalization should work)
		room := hub.GetRoom(roomData.RoomID)
		if room != nil {
			if len(room.Clients) != 0 {
				t.Errorf("Expected room to be empty after leave, but has %d clients", len(room.Clients))
			}
		}
	})
}
