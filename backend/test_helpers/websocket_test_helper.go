package test_helpers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"time"

	"turn-tracker/backend/core"
	"turn-tracker/backend/types"

	"github.com/gorilla/websocket"
)

// TestServer wraps the server for testing
type TestServer struct {
	Hub    *core.Hub
	Server *httptest.Server
}

// SetupTestServer creates a test WebSocket server
// router must be provided to avoid import cycles
func SetupTestServer(router core.MessageHandler) *TestServer {
	hub := core.NewHub()

	// Set up callbacks (minimal for testing)
	hub.OnPlayerLeft = func(roomID, clientID string, msg []byte) {}
	hub.OnTurnEnded = func(roomID string) {}

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

	return &TestServer{
		Hub:    hub,
		Server: server,
	}
}

// TestWebSocketClient wraps a WebSocket connection for testing
type TestWebSocketClient struct {
	Conn  *websocket.Conn
	Hub   *core.Hub
	MsgCh chan types.Message
	ErrCh chan error
}

// ConnectTestClient creates and connects a test WebSocket client
func ConnectTestClient(serverURL string) (*TestWebSocketClient, error) {
	u := "ws" + serverURL[4:] + "/ws"
	conn, _, err := websocket.DefaultDialer.Dial(u, nil)
	if err != nil {
		return nil, err
	}

	client := &TestWebSocketClient{
		Conn:  conn,
		MsgCh: make(chan types.Message, 10),
		ErrCh: make(chan error, 10),
	}

	// Start reading messages
	go func() {
		for {
			var msg types.Message
			err := conn.ReadJSON(&msg)
			if err != nil {
				client.ErrCh <- err
				return
			}
			client.MsgCh <- msg
		}
	}()

	return client, nil
}

// SendMessage sends a JSON message to the server
func (c *TestWebSocketClient) SendMessage(msgType string, data interface{}) error {
	msg := types.Message{
		Type: msgType,
	}

	dataBytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	msg.Data = json.RawMessage(dataBytes)

	return c.Conn.WriteJSON(msg)
}

// ReceiveMessage waits for a message with optional timeout
func (c *TestWebSocketClient) ReceiveMessage(timeout time.Duration) (types.Message, error) {
	select {
	case msg := <-c.MsgCh:
		return msg, nil
	case err := <-c.ErrCh:
		return types.Message{}, err
	case <-time.After(timeout):
		return types.Message{}, fmt.Errorf("timeout waiting for message")
	}
}

// Close closes the WebSocket connection
func (c *TestWebSocketClient) Close() error {
	return c.Conn.Close()
}

// Cleanup shuts down the test server
func (ts *TestServer) Cleanup() {
	ts.Server.Close()
}
