package startturn

import (
	"log"
	"turn-tracker/backend/core"
	"turn-tracker/backend/types"
)

// HandleStartTurn handles starting or ending a player's turn
// Uses optimistic concurrency: client sends their view of current turn, server validates
// If new_turn is empty, ends the current turn
func HandleStartTurn(hub *core.Hub, client *core.Client, expectedCurrentTurn, newTurnClientID string) {
	// If the new turn client ID is the same as the expected current turn, do nothing
	if newTurnClientID == expectedCurrentTurn {
		return
	}

	// Check if client is in a room
	if client.RoomID == "" {
		errorMsg, _ := types.NewErrorMessage("Not in a room")
		client.SafeSend(errorMsg)
		return
	}

	// Get the room
	room := hub.GetRoom(client.RoomID)
	if room == nil {
		errorMsg, _ := types.NewErrorMessage("Room not found")
		client.SafeSend(errorMsg)
		return
	}

	// States match - process the request
	// If new_turn is empty, end the current turn
	if newTurnClientID == "" {
		// Clear the current turn (validates state internally)
		room.ClearCurrentTurn()

		// Get sequence number for turn_changed message
		sequence := room.GetTurnSequence()

		// Broadcast turn ended to all players in room
		turnChangedMsg, err := NewTurnChangedMessage(client.RoomID, core.PeerInfo{}, 0, sequence)
		if err != nil {
			log.Printf("Error creating turn_changed message: %v", err)
			return
		}
		hub.BroadcastToRoom(client.RoomID, turnChangedMsg)

		log.Printf("Turn ended in room %s by client %s", client.RoomID, client.ClientID)
		return
	}

	// Try to set the new turn atomically (validates state and sets in one operation)
	if !room.SetCurrentTurn(expectedCurrentTurn, newTurnClientID) {
		// State mismatch or client not found - send state sync with current state
		currentTurnInfo := room.GetCurrentTurnInfo()
		turnStartTime := room.GetTurnStartTime()
		sequence := room.GetTurnSequence()
		turnChangedMsg, err := NewTurnChangedMessage(client.RoomID, currentTurnInfo, turnStartTime, sequence)
		if err == nil {
			// Send state sync to this client only (not broadcast)
			client.SafeSend(turnChangedMsg)
			log.Printf("Turn state mismatch for client %s in room %s: expected %s",
				client.ClientID, client.RoomID, expectedCurrentTurn)
		}
		return
	}

	// Successfully set the turn - get updated state and broadcast to all players in room
	currentTurnInfo := room.GetCurrentTurnInfo()
	turnStartTime := room.GetTurnStartTime()
	sequence := room.GetTurnSequence()
	turnChangedMsg, err := NewTurnChangedMessage(client.RoomID, currentTurnInfo, turnStartTime, sequence)
	if err != nil {
		log.Printf("Error creating turn_changed message: %v", err)
		return
	}
	hub.BroadcastToRoom(client.RoomID, turnChangedMsg)

	log.Printf("Turn started for client %s in room %s", newTurnClientID, client.RoomID)
}
