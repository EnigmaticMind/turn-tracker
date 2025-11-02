package core

import (
	"sync"
	"testing"
	"time"
)

// TestRateLimit wraps all rate limit tests
// This allows running all tests together or individually in the IDE
func TestRateLimit(t *testing.T) {
	t.Run("CheckRateLimit", func(t *testing.T) {
		t.Run("AllowsMessagesUnderLimit", func(t *testing.T) {
			client := &Client{ClientID: "test-client-1"}
			for i := 0; i < MessageRateLimit; i++ {
				if !client.CheckRateLimit() {
					t.Errorf("Expected message %d to be allowed", i+1)
				}
			}
		})

		t.Run("BlocksMessagesOverLimit", func(t *testing.T) {
			client := &Client{ClientID: "test-client-2"}
			// Send messages up to limit
			for i := 0; i < MessageRateLimit; i++ {
				client.CheckRateLimit()
			}
			// Next message should be blocked
			if client.CheckRateLimit() {
				t.Error("Expected message over limit to be blocked")
			}
		})

		t.Run("AllowsMessagesAfterWindowExpires", func(t *testing.T) {
			client := &Client{ClientID: "test-client-3"}
			// Send messages up to limit
			for i := 0; i < MessageRateLimit; i++ {
				client.CheckRateLimit()
			}
			// Should be blocked
			if client.CheckRateLimit() {
				t.Error("Expected message to be blocked at limit")
			}

			// Wait for window to expire (plus a small buffer)
			time.Sleep(MessageRateWindow + 50*time.Millisecond)

			// Should now be allowed (expired messages cleaned up in slow path)
			if !client.CheckRateLimit() {
				t.Error("Expected message to be allowed after window expires")
			}
		})

		t.Run("TracksSeparateClientsIndependently", func(t *testing.T) {
			client1 := &Client{ClientID: "test-client-4"}
			client2 := &Client{ClientID: "test-client-5"}

			// Fill up client1's limit
			for i := 0; i < MessageRateLimit; i++ {
				client1.CheckRateLimit()
			}

			// Client1 should be blocked
			if client1.CheckRateLimit() {
				t.Error("Expected client1 to be blocked")
			}

			// Client2 should still be allowed
			for i := 0; i < MessageRateLimit; i++ {
				if !client2.CheckRateLimit() {
					t.Errorf("Expected client2 message %d to be allowed", i+1)
				}
			}
		})

		t.Run("RemovesOldMessagesFromWindow", func(t *testing.T) {
			client := &Client{ClientID: "test-client-6"}
			// Send messages up to limit first
			for i := 0; i < MessageRateLimit; i++ {
				client.CheckRateLimit()
			}

			// Wait for messages to age out
			time.Sleep(MessageRateWindow + 50*time.Millisecond)

			// Should be able to send up to limit again (old messages removed in slow path)
			// First call will trigger slow path cleanup, then we can send up to limit
			if !client.CheckRateLimit() {
				t.Error("Expected message to be allowed after old messages expired (slow path cleanup)")
			}

			// Now should be able to send up to limit-1 more (already sent 1)
			for i := 0; i < MessageRateLimit-1; i++ {
				if !client.CheckRateLimit() {
					t.Errorf("Expected message %d to be allowed after cleanup", i+2)
				}
			}
		})

		t.Run("ConcurrentAccess", func(t *testing.T) {
			client := &Client{ClientID: "test-client-7"}
			var wg sync.WaitGroup
			iterations := 100
			allowed := 0
			var allowedMu sync.Mutex

			for i := 0; i < iterations; i++ {
				wg.Add(1)
				go func() {
					defer wg.Done()
					if client.CheckRateLimit() {
						allowedMu.Lock()
						allowed++
						allowedMu.Unlock()
					}
				}()
			}

			wg.Wait()

			// Should allow at most MessageRateLimit messages
			if allowed > MessageRateLimit {
				t.Errorf("Expected at most %d messages allowed, got %d", MessageRateLimit, allowed)
			}
		})

		t.Run("InitializesOnFirstCall", func(t *testing.T) {
			client := &Client{ClientID: "test-client-8"}
			// Initially rateLimit should be nil (lazy initialization)
			if client.rateLimit != nil {
				t.Error("Expected rateLimit to be nil before first call")
			}

			// First call should initialize (using sync.Once)
			client.CheckRateLimit()

			if client.rateLimit == nil {
				t.Error("Expected rateLimit to be initialized after first call")
			}

			// Subsequent calls should reuse the same instance
			initialRateLimit := client.rateLimit
			client.CheckRateLimit()
			if client.rateLimit != initialRateLimit {
				t.Error("Expected rateLimit to be reused, not recreated")
			}
		})

		t.Run("MultipleCallsMaintainState", func(t *testing.T) {
			client := &Client{ClientID: "test-client-9"}

			// Make multiple calls quickly
			for i := 0; i < 5; i++ {
				client.CheckRateLimit()
			}

			// Rate limit should track all calls (still under limit of 20)
			if !client.CheckRateLimit() {
				t.Error("Expected client to not be rate limited after 6 calls (under limit of 20)")
			}
		})

		t.Run("FastPathSkipsCleanupWhenBelowLimit", func(t *testing.T) {
			client := &Client{ClientID: "test-client-10"}

			// Send some messages below limit
			for i := 0; i < MessageRateLimit/2; i++ {
				client.CheckRateLimit()
			}

			// Wait for messages to expire
			time.Sleep(MessageRateWindow + 50*time.Millisecond)

			// Fast path doesn't clean up expired messages - it just checks count
			// So expired messages will remain until we hit the limit
			// Send more messages to fill up to limit
			for i := 0; i < MessageRateLimit/2; i++ {
				client.CheckRateLimit()
			}

			// Now at limit, next call should trigger slow path cleanup
			// Since the first half expired, cleanup should remove them
			if !client.CheckRateLimit() {
				t.Error("Expected message to be allowed after expired messages cleaned up in slow path")
			}
		})

		t.Run("SlowPathCleansUpExpiredMessages", func(t *testing.T) {
			client := &Client{ClientID: "test-client-11"}

			// Fill up to limit
			for i := 0; i < MessageRateLimit; i++ {
				client.CheckRateLimit()
			}

			// Wait for all messages to expire
			time.Sleep(MessageRateWindow + 50*time.Millisecond)

			// Next call triggers slow path which should clean up all expired messages
			// and allow the new message
			if !client.CheckRateLimit() {
				t.Error("Expected message to be allowed after slow path cleanup removes all expired messages")
			}

			// Should be able to send up to limit now
			for i := 0; i < MessageRateLimit-1; i++ {
				if !client.CheckRateLimit() {
					t.Errorf("Expected message %d to be allowed after slow path cleanup", i+2)
				}
			}
		})
	})

	t.Run("EdgeCases", func(t *testing.T) {
		t.Run("RapidMessages", func(t *testing.T) {
			client := &Client{ClientID: "rapid-client"}
			allowed := 0

			start := time.Now()
			for time.Since(start) < 100*time.Millisecond {
				if client.CheckRateLimit() {
					allowed++
				}
			}

			// Should allow at most MessageRateLimit messages per window
			if allowed > MessageRateLimit {
				t.Errorf("Expected at most %d messages in window, got %d", MessageRateLimit, allowed)
			}
		})

		t.Run("ExactLimitBoundary", func(t *testing.T) {
			client := &Client{ClientID: "boundary-client"}

			// Send exactly at the limit
			for i := 0; i < MessageRateLimit-1; i++ {
				if !client.CheckRateLimit() {
					t.Errorf("Expected message %d to be allowed (under limit)", i+1)
				}
			}

			// Last message at limit should be allowed
			if !client.CheckRateLimit() {
				t.Error("Expected message at exact limit to be allowed")
			}

			// Next message should be blocked (triggers slow path, but no expired messages to clean)
			if client.CheckRateLimit() {
				t.Error("Expected message over limit to be blocked")
			}
		})

		t.Run("WindowSliding", func(t *testing.T) {
			client := &Client{ClientID: "sliding-window-client"}

			// Send half the limit first
			for i := 0; i < MessageRateLimit/2; i++ {
				client.CheckRateLimit()
			}

			// Wait for these messages to age out (more than a full window)
			time.Sleep(MessageRateWindow + 50*time.Millisecond)

			// Fill up to limit (old messages still counted in fast path)
			for i := 0; i < MessageRateLimit/2; i++ {
				client.CheckRateLimit()
			}

			// Now at limit, next call triggers slow path which cleans up expired messages
			// After cleanup, we have MessageRateLimit/2 messages (the recent batch)
			// Should be able to send more after cleanup
			if !client.CheckRateLimit() {
				t.Error("Expected message to be allowed after slow path cleanup of expired messages")
			}

			// After cleanup and adding one more, we now have MessageRateLimit/2 + 1 messages
			// Should be able to send (MessageRateLimit - (MessageRateLimit/2 + 1)) more messages
			remaining := MessageRateLimit - (MessageRateLimit/2 + 1)
			for i := 0; i < remaining; i++ {
				if !client.CheckRateLimit() {
					t.Errorf("Expected message %d to be allowed after cleanup", i+2)
				}
			}

			// Should now be at limit
			if client.CheckRateLimit() {
				t.Error("Expected to be blocked at limit")
			}
		})
	})

	t.Run("Isolation", func(t *testing.T) {
		t.Run("ClientStateIsIndependent", func(t *testing.T) {
			client1 := &Client{ClientID: "isolate-1"}
			client2 := &Client{ClientID: "isolate-2"}
			client3 := &Client{ClientID: "isolate-3"}

			// Each client should have independent rate limits
			for i := 0; i < MessageRateLimit; i++ {
				if !client1.CheckRateLimit() {
					t.Errorf("Client1: Expected message %d to be allowed", i+1)
				}
				if !client2.CheckRateLimit() {
					t.Errorf("Client2: Expected message %d to be allowed", i+1)
				}
				if !client3.CheckRateLimit() {
					t.Errorf("Client3: Expected message %d to be allowed", i+1)
				}
			}

			// All should be blocked now
			if client1.CheckRateLimit() {
				t.Error("Expected client1 to be blocked")
			}
			if client2.CheckRateLimit() {
				t.Error("Expected client2 to be blocked")
			}
			if client3.CheckRateLimit() {
				t.Error("Expected client3 to be blocked")
			}
		})
	})
}
