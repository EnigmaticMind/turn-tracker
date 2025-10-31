# Turn Tracker Backend

A WebSocket-based real-time backend server for the Turn Tracker application. Handles room creation, player connections, and message broadcasting.

## Overview

The backend provides a pure WebSocket server written in Go that manages game rooms and facilitates real-time communication between connected clients. It replaces the previous PeerJS-based solution with a simpler, server-controlled architecture.

## Architecture

### Core Components

- **Hub** (`hub.go`): Central coordinator that manages all rooms and clients. Runs in a single goroutine for thread-safe operations.
- **Client** (`client.go`): Represents a WebSocket connection. Each client has separate goroutines for reading and writing.
- **Room** (`room.go`): Represents a game room containing multiple clients.
- **Messages** (`messages.go`): Defines the JSON message protocol for client-server communication.

### Concurrency Model

- **Single Hub Goroutine**: Processes all room/client registrations and message routing sequentially (prevents race conditions)
- **Per-Client Goroutines**: Each client connection has:
  - `readPump()`: Reads incoming messages from the WebSocket
  - `writePump()`: Writes outgoing messages to the WebSocket
- **Non-blocking Communication**: Clients and Hub communicate via channels

## Getting Started

### Prerequisites

- Go 1.21 or later
- A WebSocket client (the frontend application)

### Installation

```bash
cd backend
go mod download
```

### Running

```bash
go run .
```

The server will start on `:8080` and listen for WebSocket connections at `/ws`.

### Building

```bash
go build -o turn-tracker-backend .
./turn-tracker-backend
```

## WebSocket Endpoint

**URL**: `ws://localhost:8080/ws`

No query parameters required. All room management happens via messages.

## Message Protocol

All messages use JSON format with a `type` field and a `data` field:

```json
{
  "type": "message_type",
  "data": { ... }
}
```

### Client → Server Messages

#### Create Room

```json
{
  "type": "create_room",
  "data": {
    "room_id": "ABC123"
  }
}
```

#### Join Room

```json
{
  "type": "join_room",
  "data": {
    "room_id": "ABC123"
  }
}
```

#### Broadcast

```json
{
  "type": "broadcast",
  "data": {
    "room_id": "ABC123",
    "payload": { ... }
  }
}
```

### Server → Client Messages

#### Room Created

```json
{
  "type": "room_created",
  "data": {
    "room_id": "ABC123",
    "peers": ["client-id-1"]
  }
}
```

#### Room Joined

```json
{
  "type": "room_joined",
  "data": {
    "room_id": "ABC123",
    "peers": ["client-id-1", "client-id-2"]
  }
}
```

#### Player Joined

```json
{
  "type": "player_joined",
  "data": {
    "room_id": "ABC123",
    "peer_id": "client-id-2"
  }
}
```

#### Player Left

```json
{
  "type": "player_left",
  "data": {
    "room_id": "ABC123",
    "peer_id": "client-id-2"
  }
}
```

#### Broadcast Received

```json
{
  "type": "broadcast",
  "data": {
    "room_id": "ABC123",
    "from": "client-id-1",
    "payload": { ... }
  }
}
```

#### Error

```json
{
  "type": "error",
  "data": {
    "message": "Room not found"
  }
}
```

## Code Structure

```
backend/
├── main.go          # Server entry point, WebSocket endpoint setup
├── hub.go           # Hub struct and room/client management logic
├── client.go        # Client struct and WebSocket read/write pumps
├── room.go          # Room struct and peer listing
├── messages.go      # Message protocol definitions and constructors
└── README.md        # This file
```

### Key Functions

#### Hub Methods

- `handleCreateRoom()`: Creates a new room with explicit room ID
- `handleJoinRoom()`: Joins an existing room
- `handleBroadcast()`: Broadcasts message to all clients in a room
- `broadcastToRoom()`: Helper to send message to all room members
- `broadcastToRoomExcept()`: Send message to all room members except one

#### Client Methods

- `readPump()`: Reads messages from WebSocket, parses JSON, routes to handlers
- `writePump()`: Writes messages to WebSocket, handles ping/pong
- `handleMessage()`: Routes incoming messages by type to Hub handlers

## Features

- ✅ Explicit room creation (no auto-create)
- ✅ Room-based message broadcasting
- ✅ Automatic client ID generation
- ✅ Player join/leave notifications
- ✅ WebSocket ping/pong keepalive
- ✅ Automatic room cleanup when empty
- ✅ Thread-safe concurrent operations

## Client ID Generation

Each WebSocket connection receives a unique 16-character hexadecimal client ID (generated from 8 random bytes). This ID is used to:

- Identify message senders
- Track room creators
- List peers in rooms
- Verify broadcast permissions

## Room Lifecycle

1. **Creation**: Client sends `create_room` with a room ID (e.g., "ABC123")
2. **Joining**: Other clients send `join_room` with the same room ID
3. **Messaging**: Clients in a room can broadcast messages to all room members
4. **Cleanup**: When all clients disconnect, the room is automatically deleted

## Error Handling

- Invalid message format → `error` message sent to client
- Room not found → `error` message when trying to join
- Room already exists → `error` message when trying to create duplicate
- Not a room member → `error` message when trying to broadcast

## Performance Considerations

- **Hub Goroutine**: Single-threaded for coordination (fast map operations, typically not a bottleneck)
- **I/O Operations**: Each client handles its own WebSocket I/O in separate goroutines
- **Scalability**: Can handle thousands of concurrent connections (Go's efficient goroutines)

## Configuration

- **Port**: Hardcoded to `:8080` (can be changed in `main.go`)
- **CORS**: Currently allows all origins (`CheckOrigin: func(r *http.Request) bool { return true }`)
- **Message Size Limit**: 512KB (defined in `client.go`)
- **Read Timeout**: 60 seconds (pongWait)
- **Ping Interval**: 54 seconds (9/10 of pongWait)

## Development

### Adding New Message Types

1. Add message data struct to `messages.go`:

   ```go
   type MyNewMessageData struct {
       Field string `json:"field"`
   }
   ```

2. Add handler case in `client.go` `handleMessage()`:

   ```go
   case "my_new_message":
       var data MyNewMessageData
       json.Unmarshal(msg.Data, &data)
       // Handle message
   ```

3. Add handler method in `hub.go` if needed:
   ```go
   func (h *Hub) handleMyNewMessage(...) {
       // Implementation
   }
   ```

## Testing

Manual testing can be done using:

- Browser DevTools WebSocket console
- `wscat` CLI tool: `wscat -c ws://localhost:8080/ws`
- The frontend application

## License

See main project README for license information.
