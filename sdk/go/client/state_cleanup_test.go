package client

import (
	"testing"
	"time"
)

func TestWatchStateCleanup(t *testing.T) {
	client, err := NewClient("nats://localhost:4222", nil)
	if err != nil {
		t.Skip("Need running NATS server with JetStream")
	}
	defer client.Close()

	callCount := 0
	stop, err := client.WatchState("test.cleanup", func(state map[string]interface{}) {
		callCount++
	})
	if err != nil {
		t.Fatalf("WatchState failed: %v", err)
	}

	// Trigger a state change
	client.SetState("test.cleanup", map[string]interface{}{"value": 1})
	time.Sleep(100 * time.Millisecond)

	// Stop watching
	stop()

	// Wait a bit to ensure goroutine exits
	time.Sleep(200 * time.Millisecond)

	// Trigger another change - handler should not be called
	client.SetState("test.cleanup", map[string]interface{}{"value": 2})
	time.Sleep(100 * time.Millisecond)

	if callCount != 1 {
		t.Errorf("Expected handler called once, got %d times", callCount)
	}
}
