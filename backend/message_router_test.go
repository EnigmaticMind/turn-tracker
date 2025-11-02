package main

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/createroom"
	"turn-tracker/backend/handlers/joinroom"
	"turn-tracker/backend/handlers/startturn"
	"turn-tracker/backend/handlers/updateprofile"
	"turn-tracker/backend/test_helpers"
	"turn-tracker/backend/types"

	"github.com/gorilla/websocket"
)

// TestMessageRouter wraps all message router tests
func TestMessageRouter(t *testing.T) {
	t.Run("RoutesCreateRoom", testRoutesCreateRoom)
	t.Run("RoutesJoinRoom", testRoutesJoinRoom)
	t.Run("RoutesLeaveRoom", testRoutesLeaveRoom)
	t.Run("RoutesUpdateProfile", testRoutesUpdateProfile)
	t.Run("RoutesStartTurn", testRoutesStartTurn)
	t.Run("HandlesUnknownMessageType", testHandlesUnknownMessageType)
	t.Run("HandlesInvalidJSON", testHandlesInvalidJSON)
	t.Run("NormalizesRoomIDToUppercase", testNormalizesRoomIDToUppercase)
	t.Run("HandlesEmptyData", testHandlesEmptyData)
}

func testRoutesCreateRoom(t *testing.T) {
	server := test_helpers.SetupTestServer(messageRouter)
	defer server.Cleanup()

	client, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	err = client.SendMessage("create_room", map[string]interface{}{
		"room_id":      "",
		"display_name": "TestUser",
		"color":        "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to send create_room: %v", err)
	}

	resp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive room_created: %v", err)
	}

	if resp.Type != "room_created" {
		t.Errorf("Expected 'room_created', got '%s'", resp.Type)
	}

	var createData createroom.RoomCreatedData
	if err := json.Unmarshal(resp.Data, &createData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if createData.RoomID == "" {
		t.Error("Expected RoomID to be generated")
	}

	if len(createData.Peers) != 1 {
		t.Errorf("Expected 1 peer, got %d", len(createData.Peers))
	}

	if createData.Peers[0].DisplayName != "TestUser" {
		t.Errorf("Expected DisplayName 'TestUser', got '%s'", createData.Peers[0].DisplayName)
	}
}

func testRoutesJoinRoom(t *testing.T) {
	server := test_helpers.SetupTestServer(messageRouter)
	defer server.Cleanup()

	// Create room first
	client1, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect client1: %v", err)
	}
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	client1.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client1.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Join room
	client2, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect client2: %v", err)
	}
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	err = client2.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID,
	})
	if err != nil {
		t.Fatalf("Failed to send join_room: %v", err)
	}

	resp, err := client2.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive room_joined: %v", err)
	}

	if resp.Type != "room_joined" {
		t.Errorf("Expected 'room_joined', got '%s'", resp.Type)
	}

	var joinData joinroom.RoomJoinedData
	if err := json.Unmarshal(resp.Data, &joinData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if joinData.RoomID != roomID {
		t.Errorf("Expected RoomID '%s', got '%s'", roomID, joinData.RoomID)
	}

	if len(joinData.Peers) != 2 {
		t.Errorf("Expected 2 peers, got %d", len(joinData.Peers))
	}
}

func testRoutesLeaveRoom(t *testing.T) {
	server := test_helpers.SetupTestServer(messageRouter)
	defer server.Cleanup()

	// Create room and join
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	client1.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client1.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Leave room
	err := client1.SendMessage("leave_room", map[string]interface{}{
		"room_id": roomID,
	})
	if err != nil {
		t.Fatalf("Failed to send leave_room: %v", err)
	}

	// Should not error
	time.Sleep(200 * time.Millisecond)

	// Verify client is no longer in room
	room := server.Hub.GetRoom(roomID)
	if room != nil {
		peers := room.ListPeerInfo()
		if len(peers) != 0 {
			t.Errorf("Expected room to be empty after leave, got %d peers", len(peers))
		}
	}
}

func testRoutesUpdateProfile(t *testing.T) {
	server := test_helpers.SetupTestServer(messageRouter)
	defer server.Cleanup()

	client, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	// Create room first
	client.SendMessage("create_room", map[string]interface{}{})
	client.ReceiveMessage(5 * time.Second)

	// Update profile
	err = client.SendMessage("update_profile", map[string]interface{}{
		"display_name": "UpdatedName",
		"color":        "#00FF00",
	})
	if err != nil {
		t.Fatalf("Failed to send update_profile: %v", err)
	}

	resp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive profile_updated: %v", err)
	}

	if resp.Type != "profile_updated" {
		t.Errorf("Expected 'profile_updated', got '%s'", resp.Type)
	}

	var updateData updateprofile.ProfileUpdatedData
	if err := json.Unmarshal(resp.Data, &updateData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if updateData.DisplayName != "UpdatedName" {
		t.Errorf("Expected DisplayName 'UpdatedName', got '%s'", updateData.DisplayName)
	}

	if updateData.Color != "#00FF00" {
		t.Errorf("Expected Color '#00FF00', got '%s'", updateData.Color)
	}
}

func testRoutesStartTurn(t *testing.T) {
	server := test_helpers.SetupTestServer(messageRouter)
	defer server.Cleanup()

	// Create room
	client, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	client.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID
	clientID := createData.Peers[0].ClientID

	// Start turn
	err = client.SendMessage("start_turn", map[string]interface{}{
		"current_turn": "",
		"new_turn":     clientID,
	})
	if err != nil {
		t.Fatalf("Failed to send start_turn: %v", err)
	}

	resp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive turn_changed: %v", err)
	}

	if resp.Type != "turn_changed" {
		t.Errorf("Expected 'turn_changed', got '%s'", resp.Type)
	}

	var turnData startturn.TurnChangedData
	if err := json.Unmarshal(resp.Data, &turnData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	if turnData.RoomID != roomID {
		t.Errorf("Expected RoomID '%s', got '%s'", roomID, turnData.RoomID)
	}

	if turnData.CurrentTurn == nil {
		t.Error("Expected CurrentTurn to be set")
	}
}

func testHandlesUnknownMessageType(t *testing.T) {
	server := test_helpers.SetupTestServer(messageRouter)
	defer server.Cleanup()

	client, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	err = client.SendMessage("unknown_message_type", map[string]interface{}{
		"data": "test",
	})
	if err != nil {
		t.Fatalf("Failed to send unknown message: %v", err)
	}

	resp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive error: %v", err)
	}

	if resp.Type != "error" {
		t.Errorf("Expected 'error', got '%s'", resp.Type)
	}

	var errorData types.ErrorData
	if err := json.Unmarshal(resp.Data, &errorData); err != nil {
		t.Fatalf("Failed to unmarshal error: %v", err)
	}

	if !strings.Contains(errorData.Message, "Unknown message type") {
		t.Errorf("Expected error about unknown message type, got '%s'", errorData.Message)
	}

	if !strings.Contains(errorData.Message, "unknown_message_type") {
		t.Errorf("Expected error to contain message type, got '%s'", errorData.Message)
	}
}

func testHandlesInvalidJSON(t *testing.T) {
	server := test_helpers.SetupTestServer(messageRouter)
	defer server.Cleanup()

	client, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	// Send a message with data field that contains invalid JSON
	// We send the data as a raw JSON string (which is valid JSON),
	// but when unmarshalMessageData tries to unmarshal it into the struct,
	// it will fail if the content isn't valid JSON for that struct.
	// However, to truly test invalid JSON in data, we need to send raw bytes
	// where the data field value is not valid JSON.

	// Create a valid outer message but with data that will fail to unmarshal
	// We'll use WriteMessage to send raw bytes with malformed data
	// The trick is: we can't have invalid JSON in a JSON field, but we can send
	// a JSON string that when unmarshaled will fail.

	// Actually, the message router expects msg.Data to be json.RawMessage
	// So we need to send valid outer JSON, but the data field itself needs to be
	// something that fails when unmarshaling into JoinRoomData

	// Let's send a message where data is valid JSON but missing required fields
	// OR we can send data as a number (which is valid JSON but not a valid object)
	msgBytes := []byte(`{"type":"join_room","data":123}`)

	err = client.Conn.WriteMessage(websocket.TextMessage, msgBytes)
	if err != nil {
		t.Fatalf("Failed to send message with invalid data: %v", err)
	}

	resp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive error: %v", err)
	}

	if resp.Type != "error" {
		t.Errorf("Expected 'error', got '%s'", resp.Type)
	}

	var errorData types.ErrorData
	if err := json.Unmarshal(resp.Data, &errorData); err != nil {
		t.Fatalf("Failed to unmarshal error: %v", err)
	}

	if errorData.Message != "Invalid join_room data" {
		t.Errorf("Expected 'Invalid join_room data', got '%s'", errorData.Message)
	}
}

func testNormalizesRoomIDToUppercase(t *testing.T) {
	server := test_helpers.SetupTestServer(messageRouter)
	defer server.Cleanup()

	// Create room
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	client1.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client1.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Try to join with lowercase room ID
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	lowercaseRoomID := strings.ToLower(roomID)
	err := client2.SendMessage("join_room", map[string]interface{}{
		"room_id": lowercaseRoomID,
	})
	if err != nil {
		t.Fatalf("Failed to send join_room: %v", err)
	}

	resp, err := client2.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive room_joined: %v", err)
	}

	if resp.Type != "room_joined" {
		t.Errorf("Expected 'room_joined', got '%s'", resp.Type)
	}

	var joinData joinroom.RoomJoinedData
	if err := json.Unmarshal(resp.Data, &joinData); err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	// Should have joined successfully (normalized to uppercase)
	if joinData.RoomID != roomID {
		t.Errorf("Expected normalized RoomID '%s', got '%s'", roomID, joinData.RoomID)
	}
}

func testHandlesEmptyData(t *testing.T) {
	server := test_helpers.SetupTestServer(messageRouter)
	defer server.Cleanup()

	client, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	// Send create_room with empty data
	err = client.SendMessage("create_room", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to send create_room: %v", err)
	}

	resp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive room_created: %v", err)
	}

	// Should still work - empty data is valid for create_room
	if resp.Type != "room_created" {
		t.Errorf("Expected 'room_created', got '%s'", resp.Type)
	}
}

// TestUnmarshalMessageData tests the helper function directly
func TestUnmarshalMessageData(t *testing.T) {
	t.Run("ReturnsTrueForValidJSON", func(t *testing.T) {
		hub := core.NewHub()
		client := &core.Client{
			Hub:  hub,
			Send: make(chan []byte, 32),
		}

		msg := &types.Message{
			Type: "test",
			Data: json.RawMessage(`{"field":"value"}`),
		}

		var data struct {
			Field string `json:"field"`
		}

		result := unmarshalMessageData(msg, &data, "test", client)
		if !result {
			t.Error("Expected true for valid JSON")
		}

		if data.Field != "value" {
			t.Errorf("Expected Field 'value', got '%s'", data.Field)
		}
	})

	t.Run("ReturnsFalseForInvalidJSON", func(t *testing.T) {
		hub := core.NewHub()
		client := &core.Client{
			Hub:  hub,
			Send: make(chan []byte, 32),
		}

		msg := &types.Message{
			Type: "test",
			Data: json.RawMessage(`{invalid json`),
		}

		var data struct {
			Field string `json:"field"`
		}

		result := unmarshalMessageData(msg, &data, "test", client)
		if result {
			t.Error("Expected false for invalid JSON")
		}

		// Should have sent error message
		select {
		case errorMsg := <-client.Send:
			var msg types.Message
			if err := json.Unmarshal(errorMsg, &msg); err == nil {
				if msg.Type != "error" {
					t.Errorf("Expected error message type, got '%s'", msg.Type)
				}
				var errorData types.ErrorData
				if err := json.Unmarshal(msg.Data, &errorData); err == nil {
					if !strings.Contains(errorData.Message, "Invalid test data") {
						t.Errorf("Expected error message to contain 'Invalid test data', got '%s'", errorData.Message)
					}
				}
			}
		case <-time.After(100 * time.Millisecond):
			t.Error("Expected error message to be sent")
		}
	})

	t.Run("HandlesErrorCreatingErrorMessage", func(t *testing.T) {
		hub := core.NewHub()
		client := &core.Client{
			Hub:  hub,
			Send: make(chan []byte, 1), // Small buffer to potentially block
		}

		// Fill up the channel
		client.Send <- []byte("test")

		msg := &types.Message{
			Type: "test",
			Data: json.RawMessage(`{invalid json`),
		}

		var data struct {
			Field string `json:"field"`
		}

		// This should handle the case where SafeSend fails
		result := unmarshalMessageData(msg, &data, "test", client)
		if result {
			t.Error("Expected false for invalid JSON")
		}
	})
}
