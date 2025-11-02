package types

import (
	"encoding/json"
	"sync"
)

// Message is the wrapper struct for all WebSocket messages
type Message struct {
	Type string          `json:"type"`
	Data json.RawMessage `json:"data"`
}

// ErrorData is the data structure for error messages
type ErrorData struct {
	Message string `json:"message"`
}

var (
	// Cached error messages for common errors to avoid repeated JSON marshaling
	cachedErrors = map[string][]byte{
		"Invalid message format":    nil, // Lazy init
		"Room not found":            nil,
		"Room already exists":       nil,
		"Invalid create_room data":  nil,
		"Invalid join_room data":    nil,
		"Not a member of this room": nil,
	}
	cacheMutex sync.RWMutex
	cacheInit  sync.Once
)

// Pre-allocated string builder capacity for unknown message type errors
const unknownMsgPrefix = "Unknown message type: "

// NewErrorMessage creates an error message, using cached versions when available
// Note: Cached errors are already optimized (marshaled once), so we don't need
// to use the memory pool here. The pool is used for messages that are created
// frequently and sent via WritePump, which pools them after sending.
func NewErrorMessage(message string) ([]byte, error) {
	cacheMutex.RLock()
	cached, exists := cachedErrors[message]
	cacheMutex.RUnlock()

	// Return cached version if available
	if exists && cached != nil {
		return cached, nil
	}

	// Create new error message
	data := ErrorData{
		Message: message,
	}
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	msg := Message{
		Type: "error",
		Data: dataJSON,
	}
	result, err := json.Marshal(msg)
	if err != nil {
		return nil, err
	}

	// Cache it if it's a known common error
	if exists {
		cacheMutex.Lock()
		cachedErrors[message] = result
		cacheMutex.Unlock()
	}

	return result, nil
}

// NewUnknownMessageTypeError creates an error for unknown message types using pre-allocated buffer
func NewUnknownMessageTypeError(msgType string) ([]byte, error) {
	// Pre-allocated string concatenation - calculate exact size upfront
	buf := make([]byte, 0, len(unknownMsgPrefix)+len(msgType))
	buf = append(buf, unknownMsgPrefix...)
	buf = append(buf, msgType...)
	return NewErrorMessage(string(buf))
}
