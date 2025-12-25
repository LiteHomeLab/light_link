package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
)

// TestControlHandler_ListenForControlMessages tests that the control handler subscribes to control messages
func TestControlHandler_ListenForControlMessages(t *testing.T) {
	// Skip if NATS is not available
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("NATS not available:", err)
		return
	}
	defer nc.Close()

	handler := NewControlHandler(nc, "test-service", "192.168.1.100:aabbccddeeff:test-service")

	// Subscribe to control messages
	if err := handler.Subscribe(); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer handler.Unsubscribe()

	if !handler.IsRunning() {
		t.Error("Expected handler to be running")
	}
}

// TestControlHandler_ReceivesStopCommand tests that the service exits when receiving a stop command
func TestControlHandler_ReceivesStopCommand(t *testing.T) {
	t.Skip("Skipping test that causes process exit")

	// This test would verify that when a stop command is received,
	// the service exits with code 0
	// However, running this test would actually exit the test process
}

// TestControlHandler_ReceivesRestartCommand tests that the service exits with restart code
func TestControlHandler_ReceivesRestartCommand(t *testing.T) {
	t.Skip("Skipping test that causes process exit")

	// This test would verify that when a restart command is received,
	// the service exits with code 99
	// However, running this test would actually exit the test process
}

// TestControlHandler_IgnoresOtherInstanceMessages tests that the handler ignores messages for other instances
func TestControlHandler_IgnoresOtherInstanceMessages(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("NATS not available:", err)
		return
	}
	defer nc.Close()

	// Create a response channel to capture any logs
	// (In actual implementation, we'd use a test logger)

	handler := NewControlHandler(nc, "test-service", "192.168.1.100:aabbccddeeff:test-service")
	if err := handler.Subscribe(); err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	defer handler.Unsubscribe()

	// Send a control message for a different instance
	control := types.ControlMessage{
		Service:     "test-service",
		InstanceKey: "192.168.1.101:112233445566:test-service", // Different instance
		Command:     "stop",
		Timestamp:   time.Now().Unix(),
	}

	data, _ := json.Marshal(control)
	subject := "$LL.control.test-service"
	if err := nc.Publish(subject, data); err != nil {
		t.Fatalf("Failed to publish: %v", err)
	}

	// Give time for message to be processed
	time.Sleep(100 * time.Millisecond)

	// Handler should still be running (message was for different instance)
	if !handler.IsRunning() {
		t.Error("Handler should still be running after receiving message for different instance")
	}
}

// TestNormalizeMAC tests the MAC address normalization function
func TestNormalizeMAC(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "MAC with colons",
			input:    "aa:bb:cc:dd:ee:ff",
			expected: "aabbccddeeff",
		},
		{
			name:     "MAC without colons",
			input:    "aabbccddeeff",
			expected: "aabbccddeeff",
		},
		{
			name:     "MAC with dashes",
			input:    "aa-bb-cc-dd-ee-ff",
			expected: "aabbccddeeff",
		},
		{
			name:     "Mixed case",
			input:    "AA:BB:CC:DD:EE:FF",
			expected: "AABBCCDDEEFF",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeMAC(tt.input)
			if result != tt.expected {
				t.Errorf("normalizeMAC(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}
