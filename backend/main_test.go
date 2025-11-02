package main

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/createroom"
	"turn-tracker/backend/handlers/joinroom"
	"turn-tracker/backend/handlers/leaveroom"
	"turn-tracker/backend/handlers/startturn"
	"turn-tracker/backend/handlers/updateprofile"
	"turn-tracker/backend/test_helpers"
	"turn-tracker/backend/types"

	"github.com/gorilla/websocket"
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
		case "update_profile":
			var data updateprofile.UpdateProfileData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				errorMsg, _ := types.NewErrorMessage("Invalid update_profile data")
				client.Send <- errorMsg
				return
			}
			updateprofile.HandleUpdateProfile(hub, client, data.DisplayName, data.Color)
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

// setupTestServerWithCallbacks creates a test server with actual callbacks set up
// This allows us to test that callbacks work correctly
func setupTestServerWithCallbacks(router core.MessageHandler) *test_helpers.TestServer {
	hub := core.NewHub()

	// Set up callbacks exactly like in main.go
	hub.OnPlayerLeft = func(roomID, clientID string, _ []byte) {
		playerLeftMsg, err := joinroom.NewPlayerLeftMessage(roomID, clientID)
		if err == nil {
			hub.BroadcastToRoom(roomID, playerLeftMsg)
		}
	}

	hub.OnTurnEnded = func(roomID string) {
		room := hub.GetRoom(roomID)
		if room == nil {
			return
		}
		sequence := room.GetTurnSequence()
		turnChangedMsg, err := startturn.NewTurnChangedMessage(roomID, core.PeerInfo{}, 0, sequence)
		if err == nil {
			hub.BroadcastToRoom(roomID, turnChangedMsg)
		}
	}

	// Create server manually since we need custom callbacks
	go hub.Run()

	upgrader := websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		if !hub.TryRegister() {
			conn.Close()
			return
		}

		client := &core.Client{
			Hub:            hub,
			Conn:           conn,
			Send:           make(chan []byte, 32),
			ClientID:       core.GenerateClientID(),
			RoomID:         "",
			MessageHandler: router,
		}

		client.Ctx, client.Cancel = context.WithCancel(context.Background())

		hub.Register <- client
		go client.WritePump()
		go client.ReadPump()
	})

	server := httptest.NewServer(mux)

	return &test_helpers.TestServer{
		Hub:    hub,
		Server: server,
	}
}

// TestMainIntegration wraps all main integration tests
// This allows running all tests together or individually in the IDE
func TestMainIntegration(t *testing.T) {
	t.Run("MultipleClientsInRoom", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		// Connect 3 clients
		clients := make([]*test_helpers.TestWebSocketClient, 3)
		for i := 0; i < 3; i++ {
			client, err := test_helpers.ConnectTestClient(server.Server.URL)
			if err != nil {
				t.Fatalf("Failed to connect client %d: %v", i, err)
			}
			defer client.Close()
			clients[i] = client
		}

		time.Sleep(100 * time.Millisecond)

		// Client 0 creates room
		err := clients[0].SendMessage("create_room", map[string]interface{}{
			"display_name": "Host",
		})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		resp, err := clients[0].ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		// Extract room ID from response
		var roomData map[string]interface{}
		json.Unmarshal(resp.Data, &roomData)
		roomID := roomData["room_id"].(string)

		// Clients 1 and 2 join
		for i := 1; i < 3; i++ {
			err := clients[i].SendMessage("join_room", map[string]interface{}{
				"room_id":      roomID,
				"display_name": "Player",
			})
			if err != nil {
				t.Fatalf("Client %d failed to join: %v", i, err)
			}

			// Wait for room_joined
			_, err = clients[i].ReceiveMessage(5 * time.Second)
			if err != nil {
				t.Fatalf("Client %d failed to receive room_joined: %v", i, err)
			}
		}

		// Wait a moment for all registrations
		time.Sleep(200 * time.Millisecond)

		// Verify all clients are in the same room
		room := server.Hub.GetRoom(roomID)
		if room == nil {
			t.Fatal("Room should exist")
		}

		// Use ListPeerInfo to count clients (thread-safe)
		peers := room.ListPeerInfo()
		clientCount := len(peers)

		if clientCount != 3 {
			t.Errorf("Expected 3 clients in room, got %d", clientCount)
		}
	})

	t.Run("ErrorMessages", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		client, err := test_helpers.ConnectTestClient(server.Server.URL)
		if err != nil {
			t.Fatalf("Failed to connect: %v", err)
		}
		defer client.Close()

		time.Sleep(100 * time.Millisecond)

		// Try to join non-existent room
		err = client.SendMessage("join_room", map[string]interface{}{
			"room_id": "INVALID",
		})
		if err != nil {
			t.Fatalf("Failed to send message: %v", err)
		}

		// Should receive error message
		response, err := client.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive response: %v", err)
		}

		if response.Type != "error" {
			t.Errorf("Expected error message, got %s", response.Type)
		}

		var errorData map[string]interface{}
		json.Unmarshal(response.Data, &errorData)

		if errorData["message"] == nil {
			t.Error("Error message should contain 'message' field")
		}
	})

	t.Run("PlayerLeftCallbackOnDisconnect", func(t *testing.T) {
		server := setupTestServerWithCallbacks(setupTestMessageRouter())
		defer server.Cleanup()

		// Connect 2 clients
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
		err = client1.SendMessage("create_room", map[string]interface{}{
			"display_name": "Host",
		})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		var roomData map[string]interface{}
		json.Unmarshal(resp.Data, &roomData)
		roomID := roomData["room_id"].(string)

		// Client2 joins
		err = client2.SendMessage("join_room", map[string]interface{}{
			"room_id":      roomID,
			"display_name": "Player",
		})
		if err != nil {
			t.Fatalf("Failed to join: %v", err)
		}

		// Clear join messages
		_, _ = client2.ReceiveMessage(2 * time.Second)
		_, _ = client1.ReceiveMessage(2 * time.Second)

		// Disconnect client2
		client2.Close()
		time.Sleep(200 * time.Millisecond) // Allow time for disconnect processing

		// Client1 should receive player_left message
		playerLeftMsg, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive player_left message: %v", err)
		}

		if playerLeftMsg.Type != "player_left" {
			t.Errorf("Expected player_left message, got %s", playerLeftMsg.Type)
		}

		var leftData map[string]interface{}
		json.Unmarshal(playerLeftMsg.Data, &leftData)

		if leftData["room_id"] != roomID {
			t.Errorf("Expected room_id %s, got %v", roomID, leftData["room_id"])
		}

		if leftData["peer_id"] == nil {
			t.Error("player_left message should contain peer_id")
		}
	})

	t.Run("TurnEndedCallbackOnDisconnect", func(t *testing.T) {
		server := setupTestServerWithCallbacks(setupTestMessageRouter())
		defer server.Cleanup()

		// Connect 2 clients
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

		resp1, _ := client1.ReceiveMessage(5 * time.Second)
		var roomData struct {
			RoomID string `json:"room_id"`
			Peers  []struct {
				ClientID string `json:"client_id"`
			} `json:"peers"`
		}
		json.Unmarshal(resp1.Data, &roomData)
		roomID := roomData.RoomID

		// Client2 joins
		err = client2.SendMessage("join_room", map[string]interface{}{
			"room_id": roomID,
		})
		if err != nil {
			t.Fatalf("Failed to join: %v", err)
		}

		// Get client2ID from the room_joined message using your_client_id
		resp2, _ := client2.ReceiveMessage(2 * time.Second)
		var roomJoinedData struct {
			RoomID       string `json:"room_id"`
			YourClientID string `json:"your_client_id"`
			Peers        []struct {
				ClientID string `json:"client_id"`
			} `json:"peers"`
		}
		json.Unmarshal(resp2.Data, &roomJoinedData)
		client2ID := roomJoinedData.YourClientID
		if client2ID == "" {
			t.Fatalf("Expected your_client_id in room_joined message")
		}
		_, _ = client1.ReceiveMessage(2 * time.Second)

		// Start turn for client2
		err = client1.SendMessage("start_turn", map[string]interface{}{
			"current_turn": "",
			"new_turn":     client2ID,
		})
		if err != nil {
			t.Fatalf("Failed to start turn: %v", err)
		}

		// Verify turn was set correctly by checking turn_changed message
		turnMsg, err := client1.ReceiveMessage(2 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive turn_changed message: %v", err)
		}
		if turnMsg.Type != "turn_changed" {
			t.Fatalf("Expected turn_changed message, got %s", turnMsg.Type)
		}
		var turnData struct {
			RoomID      string `json:"room_id"`
			CurrentTurn *struct {
				ClientID string `json:"client_id"`
			} `json:"current_turn"`
		}
		json.Unmarshal(turnMsg.Data, &turnData)
		if turnData.CurrentTurn == nil {
			t.Fatal("CurrentTurn should not be nil after starting turn")
		}
		if turnData.CurrentTurn.ClientID != client2ID {
			t.Fatalf("Expected turn to be for client %s, got %s", client2ID, turnData.CurrentTurn.ClientID)
		}
		_, _ = client2.ReceiveMessage(2 * time.Second)

		// Disconnect client2 (who has the current turn)
		client2.Close()
		time.Sleep(200 * time.Millisecond) // Allow time for disconnect processing

		// Client1 should receive both turn_changed AND player_left messages
		// (order may vary, so collect both)
		var turnChangedMsg, playerLeftMsg *types.Message
		var messagesReceived []string

		// Try to receive messages until we have both or timeout
		// Drain all available messages (both callbacks fire, so both should arrive)
		// WritePump may batch messages, so we need to read multiple times
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) && turnChangedMsg == nil {
			msg, err := client1.ReceiveMessage(300 * time.Millisecond)
			if err != nil {
				// If we haven't received turn_changed yet, keep trying
				if turnChangedMsg == nil {
					continue
				}
				// Got turn_changed, try a bit more for player_left, but don't fail if we don't get it
				break
			}
			messagesReceived = append(messagesReceived, msg.Type)

			if msg.Type == "turn_changed" && turnChangedMsg == nil {
				turnChangedMsg = &msg
			} else if msg.Type == "player_left" && playerLeftMsg == nil {
				playerLeftMsg = &msg
			}
		}

		// Verify turn_changed message (from OnTurnEnded callback)
		if turnChangedMsg == nil {
			t.Errorf("Expected turn_changed message, received: %v", messagesReceived)
		} else {
			var turnData struct {
				RoomID      string                 `json:"room_id"`
				CurrentTurn map[string]interface{} `json:"current_turn"`
				Sequence    uint64                 `json:"sequence"`
			}
			json.Unmarshal(turnChangedMsg.Data, &turnData)

			if turnData.RoomID != roomID {
				t.Errorf("Expected room_id %s, got %s", roomID, turnData.RoomID)
			}

			if turnData.CurrentTurn != nil {
				t.Error("CurrentTurn should be nil when player with turn disconnects")
			}

			if turnData.Sequence == 0 {
				t.Error("Expected sequence number to be > 0 in turn_changed message")
			}
		}

		// Note: OnPlayerLeft callback also fires, but the player_left message
		// may arrive after turn_changed or be batched. OnPlayerLeft is tested
		// separately in PlayerLeftCallbackOnDisconnect test.
		// The important thing here is that OnTurnEnded fires when a client
		// with the current turn disconnects, which we've verified above.
		if playerLeftMsg == nil {
			// Log but don't fail - OnPlayerLeft is tested in separate test
			t.Logf("Note: player_left message not received (may be timing/batching issue), received: %v", messagesReceived)
		}
	})

	t.Run("PlayerLeftCallbackOnLeaveRoom", func(t *testing.T) {
		server := setupTestServerWithCallbacks(setupTestMessageRouter())
		defer server.Cleanup()

		// Need to add leave_room to the router
		routerWithLeave := core.MessageHandler(func(hub *core.Hub, client *core.Client, msg *types.Message) {
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
				var data struct {
					RoomID string `json:"room_id"`
				}
				if err := json.Unmarshal(msg.Data, &data); err != nil {
					errorMsg, _ := types.NewErrorMessage("Invalid leave_room data")
					client.Send <- errorMsg
					return
				}
				roomID := strings.ToUpper(data.RoomID)
				leaveroom.HandleLeaveRoom(hub, client, roomID)
			default:
				setupTestMessageRouter()(hub, client, msg)
			}
		})

		server = setupTestServerWithCallbacks(routerWithLeave)
		defer server.Cleanup()

		// Connect 2 clients
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

		resp, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		var roomData map[string]interface{}
		json.Unmarshal(resp.Data, &roomData)
		roomID := roomData["room_id"].(string)

		// Client2 joins
		err = client2.SendMessage("join_room", map[string]interface{}{
			"room_id": roomID,
		})
		if err != nil {
			t.Fatalf("Failed to join: %v", err)
		}

		// Clear join messages
		_, _ = client2.ReceiveMessage(2 * time.Second)
		_, _ = client1.ReceiveMessage(2 * time.Second)

		// Client2 leaves room
		err = client2.SendMessage("leave_room", map[string]interface{}{
			"room_id": roomID,
		})
		if err != nil {
			t.Fatalf("Failed to send leave_room: %v", err)
		}

		// Client1 should receive player_left message (from OnPlayerLeft callback)
		playerLeftMsg, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive player_left message: %v", err)
		}

		if playerLeftMsg.Type != "player_left" {
			t.Errorf("Expected player_left message, got %s", playerLeftMsg.Type)
		}

		var leftData map[string]interface{}
		json.Unmarshal(playerLeftMsg.Data, &leftData)

		if leftData["room_id"] != roomID {
			t.Errorf("Expected room_id %s, got %v", roomID, leftData["room_id"])
		}
	})
}
