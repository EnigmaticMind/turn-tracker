package test_helpers

import (
	"turn-tracker/backend/core"
)

// SetupTestMessageRouter creates a message router for testing
// The actual router implementation is passed as a parameter to avoid import cycles
func SetupTestMessageRouter(router core.MessageHandler) core.MessageHandler {
	return router
}

