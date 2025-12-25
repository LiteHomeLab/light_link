package client

import (
	"fmt"
	"testing"
	"time"
)

func TestWatchStateCleanup(t *testing.T) {
	config := &TLSConfig{
		CaFile:     "../../../deploy/nats/tls/ca.crt",
		CertFile:   "../../../deploy/nats/tls/demo-service.crt",
		KeyFile:    "../../../deploy/nats/tls/demo-service.key",
		ServerName: "nats-server",
	}
	c, err := NewClient("nats://172.18.200.47:4222", WithTLS(config))
	if err != nil {
		t.Skip("Need running NATS server with JetStream")
	}
	defer c.Close()

	// Use unique key name to avoid conflicts
	uniqueKey := fmt.Sprintf("test.cleanup.%d", time.Now().UnixNano())

	callCount := 0
	stop, err := c.WatchState(uniqueKey, func(state map[string]interface{}) {
		callCount++
	})
	if err != nil {
		t.Fatalf("WatchState failed: %v", err)
	}

	// Trigger a state change
	c.SetState(uniqueKey, map[string]interface{}{"value": 1})
	time.Sleep(100 * time.Millisecond)

	// Stop watching
	stop()

	// Wait a bit to ensure goroutine exits
	time.Sleep(200 * time.Millisecond)

	// Trigger another change - handler should not be called
	c.SetState(uniqueKey, map[string]interface{}{"value": 2})
	time.Sleep(100 * time.Millisecond)

	if callCount != 1 {
		t.Errorf("Expected handler called once, got %d times", callCount)
	}
}
