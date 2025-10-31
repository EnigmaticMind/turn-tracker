package types

import (
	"bytes"
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
		"Invalid broadcast data":    nil,
		"Not a member of this room": nil,
	}
	cacheMutex sync.RWMutex
	cacheInit  sync.Once
)

// Pre-allocated string builder capacity for unknown message type errors
const unknownMsgPrefix = "Unknown message type: "

// NewErrorMessage creates an error message, using cached versions when available
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

// FastParseMessageType extracts just the "type" field from JSON without full unmarshal
// Returns the type string and whether it was found
func FastParseMessageType(data []byte) (string, bool) {
	// Fast path: search for "type" field
	typeIdx := bytes.Index(data, []byte(`"type"`))
	if typeIdx == -1 {
		return "", false
	}

	// Find the colon after "type"
	colonIdx := bytes.IndexByte(data[typeIdx:], ':')
	if colonIdx == -1 {
		return "", false
	}
	colonIdx += typeIdx + 1

	// Skip whitespace
	for colonIdx < len(data) && (data[colonIdx] == ' ' || data[colonIdx] == '\t' || data[colonIdx] == '\n') {
		colonIdx++
	}

	// Check if it's a string value
	if colonIdx >= len(data) || data[colonIdx] != '"' {
		return "", false
	}
	colonIdx++ // Skip opening quote

	// Find closing quote
	endIdx := bytes.IndexByte(data[colonIdx:], '"')
	if endIdx == -1 {
		return "", false
	}
	endIdx += colonIdx

	// Extract the type string
	return string(data[colonIdx:endIdx]), true
}
