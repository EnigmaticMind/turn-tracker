package joinroom

import (
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/createroom"
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
			var data JoinRoomData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				errorMsg, _ := types.NewErrorMessage("Invalid join_room data")
				client.SafeSend(errorMsg)
				return
			}
			roomID := strings.ToUpper(data.RoomID)
			HandleJoinRoom(hub, client, roomID, data.DisplayName, data.Color)
		case "start_turn":
			var data startturn.StartTurnData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				errorMsg, _ := types.NewErrorMessage("Invalid start_turn data")
				client.SafeSend(errorMsg)
				return
			}
			startturn.HandleStartTurn(hub, client, data.CurrentTurn, data.NewTurn)
		default:
			errorMsg, _ := types.NewUnknownMessageTypeError(msg.Type)
			client.SafeSend(errorMsg)
		}
	}
}

// TestJoinRoom wraps all join_room tests
func TestJoinRoom(t *testing.T) {
	t.Run("JoinExistingRoom", testJoinExistingRoom)
	t.Run("JoinNonExistentRoom", testJoinNonExistentRoom)
	t.Run("JoinWithInvalidRoomID", testJoinWithInvalidRoomID)
	t.Run("JoinRoomWithProfile", testJoinRoomWithProfile)
	t.Run("JoinRoomWithoutProfile", testJoinRoomWithoutProfile)
	t.Run("JoinRoomAlreadyInRoom", testJoinRoomAlreadyInRoom)
	t.Run("JoinDifferentRoomFromAnother", testJoinDifferentRoomFromAnother)
	t.Run("JoinRoomReceivesPeersList", testJoinRoomReceivesPeersList)
	t.Run("JoinRoomNotifiesOthers", testJoinRoomNotifiesOthers)
	t.Run("JoinRoomWithActiveTurn", testJoinRoomWithActiveTurn)
	t.Run("JoinRoomIncludesTotalTurnTime", testJoinRoomIncludesTotalTurnTime)
	t.Run("JoinDifferentRoomWithTurnTriggersCallbacks", testJoinDifferentRoomWithTurnTriggersCallbacks)
	t.Run("JoinDifferentRoomTriggersPlayerLeftCallback", testJoinDifferentRoomTriggersPlayerLeftCallback)
	t.Run("JoinRoomWithInvalidOldRoomID", testJoinRoomWithInvalidOldRoomID)
	t.Run("JoinRoomCreatorIsLastInPeers", testJoinRoomCreatorIsLastInPeers)
}

func testJoinExistingRoom(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Client 1 creates a room
	client1, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect client1: %v", err)
	}
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	err = client1.SendMessage("create_room", map[string]interface{}{})
	if err != nil {
		t.Fatalf("Failed to send create_room: %v", err)
	}

	createResp, err := client1.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive room_created: %v", err)
	}

	var createData createroom.RoomCreatedData // Fixed: use correct type
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Client 2 joins the room
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

	joinResp, err := client2.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive room_joined: %v", err)
	}

	if joinResp.Type != "room_joined" {
		t.Errorf("Expected 'room_joined', got '%s'", joinResp.Type)
	}

	var joinData RoomJoinedData
	json.Unmarshal(joinResp.Data, &joinData)

	if joinData.RoomID != roomID {
		t.Errorf("Expected RoomID '%s', got '%s'", roomID, joinData.RoomID)
	}

	if len(joinData.Peers) != 2 {
		t.Errorf("Expected 2 peers, got %d", len(joinData.Peers))
	}

	// Client 1 should receive player_joined notification
	playerJoined, err := client1.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive player_joined: %v", err)
	}

	if playerJoined.Type != "player_joined" {
		t.Errorf("Expected 'player_joined', got '%s'", playerJoined.Type)
	}
}

func testJoinNonExistentRoom(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	client, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	// Use a valid format but non-existent room ID (4 uppercase letters)
	err = client.SendMessage("join_room", map[string]interface{}{
		"room_id": "ABCD",
	})
	if err != nil {
		t.Fatalf("Failed to send join_room: %v", err)
	}

	resp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive error: %v", err)
	}

	if resp.Type != "error" {
		t.Errorf("Expected 'error', got '%s'", resp.Type)
	}

	var errorData types.ErrorData
	json.Unmarshal(resp.Data, &errorData)

	if errorData.Message != "Room not found" {
		t.Errorf("Expected error 'Room not found', got '%s'", errorData.Message)
	}
}

func testJoinWithInvalidRoomID(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	client, err := test_helpers.ConnectTestClient(server.Server.URL)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	err = client.SendMessage("join_room", map[string]interface{}{
		"room_id": "XX", // Too short
	})
	if err != nil {
		t.Fatalf("Failed to send join_room: %v", err)
	}

	resp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive error: %v", err)
	}

	if resp.Type != "error" {
		t.Errorf("Expected 'error', got '%s'", resp.Type)
	}

	var errorData types.ErrorData
	json.Unmarshal(resp.Data, &errorData)

	if errorData.Message != "Invalid game ID format" {
		t.Errorf("Expected error 'Invalid game ID format', got '%s'", errorData.Message)
	}
}

func testJoinRoomWithProfile(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Create room
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)
	client1.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client1.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Join with profile
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	err := client2.SendMessage("join_room", map[string]interface{}{
		"room_id":      roomID,
		"display_name": "TestUser",
		"color":        "#FF5733",
	})
	if err != nil {
		t.Fatalf("Failed to send: %v", err)
	}

	joinResp, err := client2.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive: %v", err)
	}

	var joinData RoomJoinedData
	json.Unmarshal(joinResp.Data, &joinData)

	// Find client2 in peers list
	var client2Peer *core.PeerInfo
	for i := range joinData.Peers {
		if joinData.Peers[i].DisplayName == "TestUser" {
			client2Peer = &joinData.Peers[i]
			break
		}
	}

	if client2Peer == nil {
		t.Fatal("Client2 not found in peers list")
	}

	if client2Peer.DisplayName != "TestUser" {
		t.Errorf("Expected DisplayName 'TestUser', got '%s'", client2Peer.DisplayName)
	}

	if client2Peer.Color != "#FF5733" {
		t.Errorf("Expected Color '#FF5733', got '%s'", client2Peer.Color)
	}
}

func testJoinRoomWithoutProfile(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Create room
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)
	client1.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client1.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Join without profile
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	client2.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID,
	})

	joinResp, _ := client2.ReceiveMessage(5 * time.Second)
	var joinData RoomJoinedData
	json.Unmarshal(joinResp.Data, &joinData)

	// Server should generate profile
	var client2Peer *core.PeerInfo
	for i := range joinData.Peers {
		if joinData.Peers[i].DisplayName != createData.Peers[0].DisplayName {
			client2Peer = &joinData.Peers[i]
			break
		}
	}

	if client2Peer == nil {
		t.Fatal("Client2 not found in peers list")
	}

	if client2Peer.DisplayName == "" {
		t.Error("DisplayName should be generated by server")
	}

	if client2Peer.Color == "" {
		t.Error("Color should be generated by server")
	}
}

func testJoinRoomAlreadyInRoom(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	client, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	// Create room
	client.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Try to join the same room again
	client.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID,
	})

	resp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive: %v", err)
	}

	if resp.Type != "error" {
		t.Errorf("Expected 'error', got '%s'", resp.Type)
	}

	var errorData types.ErrorData
	json.Unmarshal(resp.Data, &errorData)

	if errorData.Message != "Already in this room" {
		t.Errorf("Expected 'Already in this room', got '%s'", errorData.Message)
	}
}

func testJoinDifferentRoomFromAnother(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Create room 1
	client, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	client.SendMessage("create_room", map[string]interface{}{})
	createResp1, _ := client.ReceiveMessage(5 * time.Second)
	var createData1 createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp1.Data, &createData1)
	roomID1 := createData1.RoomID

	// Create room 2 with another client
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	client2.SendMessage("create_room", map[string]interface{}{})
	createResp2, _ := client2.ReceiveMessage(5 * time.Second)
	var createData2 createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp2.Data, &createData2)
	roomID2 := createData2.RoomID

	// Client 1 joins room 2 (should leave room 1)
	client.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID2,
	})

	joinResp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive room_joined: %v", err)
	}

	if joinResp.Type != "room_joined" {
		t.Errorf("Expected 'room_joined', got '%s'", joinResp.Type)
	}

	var joinData RoomJoinedData
	json.Unmarshal(joinResp.Data, &joinData)

	if joinData.RoomID != roomID2 {
		t.Errorf("Expected RoomID '%s', got '%s'", roomID2, joinData.RoomID)
	}

	// Verify client is in room 2 (should have 2 peers now)
	if len(joinData.Peers) != 2 {
		t.Errorf("Expected 2 peers in room 2, got %d", len(joinData.Peers))
	}

	// Verify client left room 1 (room 1 should now be empty)
	room1 := server.Hub.GetRoom(roomID1)
	if room1 != nil {
		peers := room1.ListPeerInfo()
		if len(peers) != 0 {
			t.Errorf("Expected room 1 to be empty, but has %d peers", len(peers))
		}
	}
}

func testJoinRoomReceivesPeersList(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Create room and add 3 clients
	var roomID string
	clients := make([]*test_helpers.TestWebSocketClient, 3)

	for i := 0; i < 3; i++ {
		client, _ := test_helpers.ConnectTestClient(server.Server.URL)
		clients[i] = client
		defer client.Close()
		time.Sleep(100 * time.Millisecond)

		if i == 0 {
			client.SendMessage("create_room", map[string]interface{}{})
			resp, _ := client.ReceiveMessage(5 * time.Second)
			var data createroom.RoomCreatedData // Fixed
			json.Unmarshal(resp.Data, &data)
			roomID = data.RoomID
		} else {
			client.SendMessage("join_room", map[string]interface{}{
				"room_id": roomID,
			})
			// Wait for room_joined message
			resp, _ := client.ReceiveMessage(5 * time.Second)
			if resp.Type != "room_joined" {
				t.Fatalf("Expected room_joined, got %s", resp.Type)
			}
			var joinData RoomJoinedData
			json.Unmarshal(resp.Data, &joinData)
			if i == 2 {
				// Last client's response should have all 3 peers
				if len(joinData.Peers) != 3 {
					t.Errorf("Expected 3 peers, got %d", len(joinData.Peers))
				}
			}
		}
	}
}

func testJoinRoomNotifiesOthers(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Client 1 creates room
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	client1.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client1.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Client 2 joins
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	client2.SendMessage("join_room", map[string]interface{}{
		"room_id":      roomID,
		"display_name": "Joiner",
	})

	// Client 2 receives room_joined
	client2.ReceiveMessage(5 * time.Second)

	// Client 1 should receive player_joined
	notif, err := client1.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive player_joined: %v", err)
	}

	if notif.Type != "player_joined" {
		t.Errorf("Expected 'player_joined', got '%s'", notif.Type)
	}

	var playerData PlayerJoinedData
	json.Unmarshal(notif.Data, &playerData)

	if playerData.DisplayName != "Joiner" {
		t.Errorf("Expected DisplayName 'Joiner', got '%s'", playerData.DisplayName)
	}

	if playerData.RoomID != roomID {
		t.Errorf("Expected RoomID '%s', got '%s'", roomID, playerData.RoomID)
	}
}

func testJoinRoomWithActiveTurn(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Create room with active turn
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	client1.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client1.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Start a turn
	client1.SendMessage("start_turn", map[string]interface{}{
		"new_turn": createData.Peers[0].ClientID,
	})
	client1.ReceiveMessage(5 * time.Second) // Wait for turn_changed

	// Client 2 joins - should see active turn
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	client2.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID,
	})

	joinResp, err := client2.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive: %v", err)
	}

	var joinData RoomJoinedData
	json.Unmarshal(joinResp.Data, &joinData)

	if joinData.CurrentTurn == nil {
		t.Error("Expected CurrentTurn to be set, got nil")
	}

	if joinData.CurrentTurn.ClientID != createData.Peers[0].ClientID {
		t.Errorf("Expected CurrentTurn ClientID '%s', got '%s'", createData.Peers[0].ClientID, joinData.CurrentTurn.ClientID)
	}
}

func testJoinRoomIncludesTotalTurnTime(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Client 1 creates room
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	client1.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client1.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID

	// Client 2 joins
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	client2.SendMessage("join_room", map[string]interface{}{
		"room_id":      roomID,
		"display_name": "Joiner",
	})

	// Client 2 receives room_joined
	client2.ReceiveMessage(5 * time.Second)

	// Client 1 should receive player_joined with TotalTurnTime
	notif, err := client1.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive player_joined: %v", err)
	}

	var playerData PlayerJoinedData
	json.Unmarshal(notif.Data, &playerData)

	// TotalTurnTime should be present (0 for new player)
	if playerData.TotalTurnTime != 0 {
		t.Errorf("Expected TotalTurnTime 0 for new player, got %d", playerData.TotalTurnTime)
	}

	// Verify it's in the field (not missing)
	notifJSON, _ := json.Marshal(notif.Data)
	if !strings.Contains(string(notifJSON), "total_turn_time") {
		t.Error("player_joined message should include total_turn_time field")
	}
}

func testJoinDifferentRoomWithTurnTriggersCallbacks(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	turnEndedCalled := false
	var turnEndedRoomID string
	var mu sync.Mutex

	hub := server.Hub
	hub.OnTurnEnded = func(roomID string) {
		mu.Lock()
		turnEndedCalled = true
		turnEndedRoomID = roomID
		mu.Unlock()
	}
	hub.OnPlayerLeft = func(roomID, clientID string, msg []byte) {}

	// Create room 1
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	client1.SendMessage("create_room", map[string]interface{}{})
	createResp1, _ := client1.ReceiveMessage(5 * time.Second)
	var createData1 createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp1.Data, &createData1)
	roomID1 := createData1.RoomID

	// Add another client to room 1 (so room isn't empty when client1 leaves)
	client3, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client3.Close()
	time.Sleep(100 * time.Millisecond)

	client3.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID1,
	})
	client3.ReceiveMessage(5 * time.Second) // Wait for room_joined

	// Start a turn in room 1
	client1.SendMessage("start_turn", map[string]interface{}{
		"new_turn": createData1.Peers[0].ClientID,
	})
	client1.ReceiveMessage(5 * time.Second) // Wait for turn_changed

	// Create room 2
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	client2.SendMessage("create_room", map[string]interface{}{})
	createResp2, _ := client2.ReceiveMessage(5 * time.Second)
	var createData2 createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp2.Data, &createData2)
	roomID2 := createData2.RoomID

	// Client 1 joins room 2 (should trigger OnTurnEnded for room 1)
	client1.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID2,
	})

	client1.ReceiveMessage(5 * time.Second) // Wait for room_joined
	time.Sleep(200 * time.Millisecond)      // Wait for callbacks to fire

	mu.Lock()
	wasCalled := turnEndedCalled
	roomID := turnEndedRoomID
	mu.Unlock()

	if !wasCalled {
		t.Error("Expected OnTurnEnded callback to be called when leaving room with active turn")
	}

	if roomID != roomID1 {
		t.Errorf("Expected OnTurnEnded to be called with roomID '%s', got '%s'", roomID1, roomID)
	}
}

func testJoinDifferentRoomTriggersPlayerLeftCallback(t *testing.T) {
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

	// Create room 1
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	client1.SendMessage("create_room", map[string]interface{}{})
	createResp1, _ := client1.ReceiveMessage(5 * time.Second)
	var createData1 createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp1.Data, &createData1)
	roomID1 := createData1.RoomID

	// Add client 2 to room 1
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	client2.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID1,
	})
	client2.ReceiveMessage(5 * time.Second) // Wait for room_joined

	// Create room 2
	client3, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client3.Close()
	time.Sleep(100 * time.Millisecond)

	client3.SendMessage("create_room", map[string]interface{}{})
	createResp2, _ := client3.ReceiveMessage(5 * time.Second)
	var createData2 createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp2.Data, &createData2)
	roomID2 := createData2.RoomID

	// Client 2 joins room 2 (should trigger OnPlayerLeft for room 1)
	client2.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID2,
	})

	client2.ReceiveMessage(5 * time.Second) // Wait for room_joined
	time.Sleep(100 * time.Millisecond)      // Wait for callbacks

	if !playerLeftCalled {
		t.Error("Expected OnPlayerLeft callback to be called when leaving room")
	}

	if leftRoomID != roomID1 {
		t.Errorf("Expected OnPlayerLeft to be called with roomID '%s', got '%s'", roomID1, leftRoomID)
	}

	// Verify client ID matches
	if leftClientID == "" {
		t.Error("Expected OnPlayerLeft to be called with clientID, got empty string")
	}
}

func testJoinRoomWithInvalidOldRoomID(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Create a client and have them create/join a room
	client, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client.Close()
	time.Sleep(100 * time.Millisecond)

	// Create room 1
	client.SendMessage("create_room", map[string]interface{}{})
	createResp1, _ := client.ReceiveMessage(5 * time.Second)
	var createData1 createroom.RoomCreatedData
	json.Unmarshal(createResp1.Data, &createData1)
	roomID1 := createData1.RoomID

	// Verify client is in room 1
	room1 := server.Hub.GetRoom(roomID1)
	if room1 == nil {
		t.Fatal("Room 1 should exist")
	}

	// Now manually delete room 1 from hub to simulate it being deleted
	// (This tests the code path where oldRoom == nil in HandleJoinRoom)
	server.Hub.DeleteRoom(roomID1)

	// Verify room is deleted
	if server.Hub.GetRoom(roomID1) != nil {
		t.Fatal("Room 1 should be deleted")
	}

	// Create a new valid room
	client2, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client2.Close()
	time.Sleep(100 * time.Millisecond)

	client2.SendMessage("create_room", map[string]interface{}{})
	createResp2, _ := client2.ReceiveMessage(5 * time.Second)
	var createData2 createroom.RoomCreatedData
	json.Unmarshal(createResp2.Data, &createData2)
	roomID2 := createData2.RoomID

	// Try to join room 2 - should handle invalid old RoomID gracefully
	// (client.RoomID is still set to roomID1, but that room no longer exists)
	client.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID2,
	})

	joinResp, err := client.ReceiveMessage(5 * time.Second)
	if err != nil {
		t.Fatalf("Failed to receive: %v", err)
	}

	// Should succeed - invalid old room ID is cleared and join proceeds
	if joinResp.Type != "room_joined" {
		t.Errorf("Expected 'room_joined', got '%s'", joinResp.Type)
	}

	var joinData RoomJoinedData
	json.Unmarshal(joinResp.Data, &joinData)

	if joinData.RoomID != roomID2 {
		t.Errorf("Expected RoomID '%s', got '%s'", roomID2, joinData.RoomID)
	}
}

func testJoinRoomCreatorIsLastInPeers(t *testing.T) {
	server := test_helpers.SetupTestServer(setupTestMessageRouter())
	defer server.Cleanup()

	// Create room
	client1, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client1.Close()
	time.Sleep(100 * time.Millisecond)

	client1.SendMessage("create_room", map[string]interface{}{})
	createResp, _ := client1.ReceiveMessage(5 * time.Second)
	var createData createroom.RoomCreatedData // Fixed
	json.Unmarshal(createResp.Data, &createData)
	roomID := createData.RoomID
	creatorID := createData.Peers[0].ClientID

	// Add 2 more clients
	for i := 0; i < 2; i++ {
		client, _ := test_helpers.ConnectTestClient(server.Server.URL)
		defer client.Close()
		time.Sleep(100 * time.Millisecond)
		client.SendMessage("join_room", map[string]interface{}{
			"room_id": roomID,
		})
		client.ReceiveMessage(5 * time.Second)
	}

	// Last client should see creator as last in peers list
	client3, _ := test_helpers.ConnectTestClient(server.Server.URL)
	defer client3.Close()
	time.Sleep(100 * time.Millisecond)

	client3.SendMessage("join_room", map[string]interface{}{
		"room_id": roomID,
	})

	joinResp, _ := client3.ReceiveMessage(5 * time.Second)
	var joinData RoomJoinedData
	json.Unmarshal(joinResp.Data, &joinData)

	if len(joinData.Peers) != 4 {
		t.Fatalf("Expected 4 peers, got %d", len(joinData.Peers))
	}

	// Creator should be last
	lastPeer := joinData.Peers[len(joinData.Peers)-1]
	if lastPeer.ClientID != creatorID {
		t.Errorf("Expected creator '%s' to be last, but last peer is '%s'", creatorID, lastPeer.ClientID)
	}
}
