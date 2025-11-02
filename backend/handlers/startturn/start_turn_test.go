package startturn

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"turn-tracker/backend/core"
	"turn-tracker/backend/handlers/createroom"
	"turn-tracker/backend/handlers/joinroom"
	"turn-tracker/backend/handlers/updateprofile"
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
		case "update_profile":
			var data updateprofile.UpdateProfileData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				errorMsg, _ := types.NewErrorMessage("Invalid update_profile data")
				client.Send <- errorMsg
				return
			}
			updateprofile.HandleUpdateProfile(hub, client, data.DisplayName, data.Color)
		case "start_turn":
			var data StartTurnData
			if err := json.Unmarshal(msg.Data, &data); err != nil {
				errorMsg, _ := types.NewErrorMessage("Invalid start_turn data")
				client.Send <- errorMsg
				return
			}
			HandleStartTurn(hub, client, data.CurrentTurn, data.NewTurn)
		default:
			errorMsg, _ := types.NewUnknownMessageTypeError(msg.Type)
			client.Send <- errorMsg
		}
	}
}

// TestStartTurn wraps all start_turn tests
// This allows running all tests together or individually in the IDE
func TestStartTurn(t *testing.T) {
	t.Run("StartTurnMessageFormat", func(t *testing.T) {
		server := test_helpers.SetupTestServer(setupTestMessageRouter())
		defer server.Cleanup()

		// Create two clients
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
			"display_name": "Player1",
			"color":        "#FF5733",
		})
		if err != nil {
			t.Fatalf("Failed to create room: %v", err)
		}

		resp1, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_created: %v", err)
		}

		var roomData struct {
			RoomID string `json:"room_id"`
			Peers  []struct {
				ClientID string `json:"client_id"`
			} `json:"peers"`
		}
		if err := json.Unmarshal(resp1.Data, &roomData); err != nil {
			t.Fatalf("Failed to unmarshal room_created: %v", err)
		}
		roomID := roomData.RoomID

		// Client2 joins room
		err = client2.SendMessage("join_room", map[string]interface{}{
			"room_id":      roomID,
			"display_name": "Player2",
			"color":        "#33FF57",
		})
		if err != nil {
			t.Fatalf("Failed to join room: %v", err)
		}

		// Client2 receives room_joined
		roomJoined, err := client2.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_joined: %v", err)
		}

		// Extract client2ID from room_joined peers list
		var roomJoinedData struct {
			Peers []struct {
				ClientID string `json:"client_id"`
			} `json:"peers"`
		}
		if err := json.Unmarshal(roomJoined.Data, &roomJoinedData); err != nil {
			t.Fatalf("Failed to unmarshal room_joined: %v", err)
		}

		// Find client2ID from peers (creator is last, so client2 should be first)
		var client2ID string
		if len(roomJoinedData.Peers) > 1 {
			// Creator is last, so joiner (client2) is first
			client2ID = roomJoinedData.Peers[0].ClientID
		} else {
			// Fallback: get from player_joined notification
			playerJoined, err := client1.ReceiveMessage(5 * time.Second)
			if err != nil {
				t.Fatalf("Failed to receive player_joined: %v", err)
			}

			var joinData struct {
				PeerID string `json:"peer_id"`
			}
			if err := json.Unmarshal(playerJoined.Data, &joinData); err != nil {
				t.Fatalf("Failed to unmarshal player_joined: %v", err)
			}
			client2ID = joinData.PeerID
		}

		if client2ID == "" {
			t.Fatal("Could not determine client2ID")
		}

		// Client1 starts turn for client2
		err = client1.SendMessage("start_turn", map[string]interface{}{
			"current_turn": "",
			"new_turn":     client2ID,
		})
		if err != nil {
			t.Fatalf("Failed to start turn: %v", err)
		}

		// Both clients should receive turn_changed
		// Client1 might receive player_joined first, so skip until we get turn_changed
		var turnResp1 types.Message
		for {
			msg, err := client1.ReceiveMessage(5 * time.Second)
			if err != nil {
				t.Fatalf("Client1 failed to receive turn_changed: %v", err)
			}
			if msg.Type == "turn_changed" {
				turnResp1 = msg
				break
			}
		}

		turnResp2, err := client2.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Client2 failed to receive turn_changed: %v", err)
		}
		if turnResp2.Type != "turn_changed" {
			t.Fatalf("Client2 received unexpected message type: %s", turnResp2.Type)
		}

		// Validate turn_changed format - Data is already json.RawMessage
		var turnData1, turnData2 TurnChangedData
		if err := json.Unmarshal(turnResp1.Data, &turnData1); err != nil {
			t.Fatalf("Failed to unmarshal turn_changed data (client1): %v", err)
		}
		if err := json.Unmarshal(turnResp2.Data, &turnData2); err != nil {
			t.Fatalf("Failed to unmarshal turn_changed data (client2): %v", err)
		}

		if turnData1.RoomID != roomID {
			t.Errorf("Expected RoomID %s, got %s", roomID, turnData1.RoomID)
		}
		if turnData1.CurrentTurn == nil {
			t.Errorf("CurrentTurn should not be nil after starting turn. Message: %s", string(turnResp1.Data))
		} else {
			if turnData1.CurrentTurn.ClientID != client2ID {
				t.Errorf("Expected CurrentTurn.ClientID %s, got %s", client2ID, turnData1.CurrentTurn.ClientID)
			}
		}
		if turnData1.TurnStartTime == nil {
			t.Errorf("TurnStartTime should not be nil after starting turn. Message: %s", string(turnResp1.Data))
		}

		// Both should receive identical data
		if turnData1.RoomID != turnData2.RoomID {
			t.Error("Both clients should receive same room_id")
		}
		if turnData1.CurrentTurn != nil && turnData2.CurrentTurn != nil {
			if turnData1.CurrentTurn.ClientID != turnData2.CurrentTurn.ClientID {
				t.Error("Both clients should receive same current_turn")
			}
			if turnData1.TurnStartTime != nil && turnData2.TurnStartTime != nil {
				if *turnData1.TurnStartTime != *turnData2.TurnStartTime {
					t.Error("Both clients should receive same turn_start_time")
				}
			}
		}
	})

	t.Run("EndTurnMessageFormat", func(t *testing.T) {
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

		// Setup room with two clients
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

		err = client2.SendMessage("join_room", map[string]interface{}{
			"room_id": roomData.RoomID,
		})
		if err != nil {
			t.Fatalf("Failed to join room: %v", err)
		}

		// Client2 receives room_joined
		roomJoined, err := client2.ReceiveMessage(2 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive room_joined: %v", err)
		}

		// Client1 receives player_joined notification
		playerJoined, err := client1.ReceiveMessage(2 * time.Second)
		if err != nil {
			t.Fatalf("Failed to receive player_joined: %v", err)
		}

		// Extract client2ID (creator is last, so joiner is first)
		var client2ID string
		var joinResp struct {
			Peers []struct {
				ClientID string `json:"client_id"`
			} `json:"peers"`
		}
		if err := json.Unmarshal(roomJoined.Data, &joinResp); err == nil && len(joinResp.Peers) > 1 {
			// Creator is last, so joiner (client2) is first
			client2ID = joinResp.Peers[0].ClientID
		} else {
			// Fallback: get from player_joined
			var pj struct {
				PeerID string `json:"peer_id"`
			}
			if err := json.Unmarshal(playerJoined.Data, &pj); err != nil {
				t.Fatalf("Failed to unmarshal player_joined: %v", err)
			}
			client2ID = pj.PeerID
		}

		if client2ID == "" {
			t.Fatal("Could not determine client2ID")
		}

		// Start a turn
		err = client1.SendMessage("start_turn", map[string]interface{}{
			"current_turn": "",
			"new_turn":     client2ID,
		})
		if err != nil {
			t.Fatalf("Failed to start turn: %v", err)
		}

		// Receive turn_changed messages
		_, err = client1.ReceiveMessage(2 * time.Second)
		if err != nil {
			t.Fatalf("Client1 failed to receive turn_changed after start: %v", err)
		}
		_, err = client2.ReceiveMessage(2 * time.Second)
		if err != nil {
			t.Fatalf("Client2 failed to receive turn_changed after start: %v", err)
		}

		// End the turn
		err = client1.SendMessage("start_turn", map[string]interface{}{
			"current_turn": client2ID,
			"new_turn":     "",
		})
		if err != nil {
			t.Fatalf("Failed to end turn: %v", err)
		}

		// Both should receive turn_changed with null current_turn
		turnResp1, err := client1.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Client1 failed to receive turn_changed: %v", err)
		}

		turnResp2, err := client2.ReceiveMessage(5 * time.Second)
		if err != nil {
			t.Fatalf("Client2 failed to receive turn_changed: %v", err)
		}

		var turnData1, turnData2 TurnChangedData
		if err := json.Unmarshal(turnResp1.Data, &turnData1); err != nil {
			t.Fatalf("Failed to unmarshal turn_changed (client1): %v", err)
		}
		if err := json.Unmarshal(turnResp2.Data, &turnData2); err != nil {
			t.Fatalf("Failed to unmarshal turn_changed (client2): %v", err)
		}

		if turnData1.CurrentTurn != nil {
			t.Error("CurrentTurn should be nil after ending turn")
		}
		if turnData1.TurnStartTime != nil {
			t.Error("TurnStartTime should be nil after ending turn")
		}

		// Both should receive identical null data
		if turnData1.RoomID != turnData2.RoomID {
			t.Error("Both clients should receive same room_id")
		}
	})
}
