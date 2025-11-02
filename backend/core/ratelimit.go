package core

import (
	"sync"
	"time"
)

const (
	// Rate limiting constants
	MessageRateLimit  = 20          // messages per window
	MessageRateWindow = time.Second // time window
)

type clientRateLimit struct {
	messages []time.Time
	mu       sync.Mutex
}

// CheckRateLimit checks if the client has exceeded the rate limit
// Returns true if allowed, false if rate limited
// Rate limit state is tied to the Client's lifecycle - automatically cleaned up when client is deleted
func (c *Client) CheckRateLimit() bool {
	c.rateLimitOnce.Do(func() {
		c.rateLimit = &clientRateLimit{
			messages: make([]time.Time, 0, MessageRateLimit),
		}
	})

	c.rateLimit.mu.Lock()
	defer c.rateLimit.mu.Unlock()

	now := time.Now()
	messages := c.rateLimit.messages

	// Fast path: Below limit, just add and return (skip cleanup)
	if len(messages) < MessageRateLimit {
		c.rateLimit.messages = append(messages, now)
		return true
	}

	// Slow path: At/above limit, need to clean up expired messages first
	windowStart := now.Add(-MessageRateWindow)

	// Find first unexpired message (messages are chronological, oldest first)
	// Once we find one unexpired, all after it are also unexpired
	firstValidIndex := len(messages)
	for i := 0; i < len(messages); i++ {
		if messages[i].After(windowStart) {
			firstValidIndex = i
			break
		}
	}

	// Clean up expired messages (remove everything before firstValidIndex)
	if firstValidIndex < len(messages) {
		c.rateLimit.messages = messages[firstValidIndex:]
	} else {
		// All messages expired
		c.rateLimit.messages = messages[:0]
	}

	// Check limit after cleanup
	if len(c.rateLimit.messages) >= MessageRateLimit {
		return false
	}

	// Add current message
	c.rateLimit.messages = append(c.rateLimit.messages, now)
	return true
}
