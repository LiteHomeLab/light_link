package client

import (
    "testing"
    "time"
)

func TestSetGetState(t *testing.T) {
    config := &TLSConfig{
        CaFile:     "../../../deploy/nats/tls/ca.crt",
        CertFile:   "../../../deploy/nats/tls/demo-service.crt",
        KeyFile:    "../../../deploy/nats/tls/demo-service.key",
        ServerName: "nats-server",
    }
    c, err := NewClient("nats://172.18.200.47:4222", WithTLS(config))
    if err != nil {
        t.Skip("Need running NATS server with JetStream:", err)
    }
    defer c.Close()

    // Set state
    err = c.SetState("test.key", map[string]interface{}{"value": 123})
    if err != nil {
        t.Fatalf("SetState failed: %v", err)
    }

    // Get state
    state, err := c.GetState("test.key")
    if err != nil {
        t.Fatalf("GetState failed: %v", err)
    }

    if state["value"].(float64) != 123 {
        t.Errorf("Expected 123, got %v", state["value"])
    }
}

func TestWatchState(t *testing.T) {
    config := &TLSConfig{
        CaFile:     "../../../deploy/nats/tls/ca.crt",
        CertFile:   "../../../deploy/nats/tls/demo-service.crt",
        KeyFile:    "../../../deploy/nats/tls/demo-service.key",
        ServerName: "nats-server",
    }
    c, err := NewClient("nats://172.18.200.47:4222", WithTLS(config))
    if err != nil {
        t.Skip("Need running NATS server with JetStream:", err)
    }
    defer c.Close()

    changes := make(chan map[string]interface{}, 1)

    // Watch state changes
    stop, err := c.WatchState("test.watch", func(state map[string]interface{}) {
        changes <- state
    })
    if err != nil {
        t.Fatalf("WatchState failed: %v", err)
    }
    defer stop()

    // Modify state
    c.SetState("test.watch", map[string]interface{}{"status": "updated"})

    // Wait for notification
    select {
    case state := <-changes:
        if state["status"] != "updated" {
            t.Errorf("Expected 'updated', got '%v'", state["status"])
        }
    case <-time.After(2 * time.Second):
        t.Error("Timeout waiting for state change")
    }
}
