package client

import (
    "testing"
    "time"
)

func TestCall(t *testing.T) {
    // This test requires a running NATS server and service
    config := &TLSConfig{
        CaFile:     "../../../deploy/nats/tls/ca.crt",
        CertFile:   "../../../deploy/nats/tls/demo-service.crt",
        KeyFile:    "../../../deploy/nats/tls/demo-service.key",
        ServerName: "nats-server",
    }
    client, err := NewClient("nats://172.18.200.47:4222", config)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }
    defer client.Close()

    // Test calling non-existent service
    result, err := client.Call("test-service", "testMethod", map[string]interface{}{"key": "value"})
    if err == nil {
        t.Error("Expected error for non-existent service")
    }
    t.Logf("Expected error: %v", err)
    t.Logf("Result: %v", result)
}

func TestCallWithTimeout(t *testing.T) {
    config := &TLSConfig{
        CaFile:     "../../../deploy/nats/tls/ca.crt",
        CertFile:   "../../../deploy/nats/tls/demo-service.crt",
        KeyFile:    "../../../deploy/nats/tls/demo-service.key",
        ServerName: "nats-server",
    }
    client, err := NewClient("nats://172.18.200.47:4222", config)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }
    defer client.Close()

    // Test timeout
    done := make(chan bool)
    go func() {
        _, _ = client.Call("timeout-service", "slowMethod", nil)
        done <- true
    }()

    select {
    case <-done:
        t.Log("Call completed")
    case <-time.After(2 * time.Second):
        t.Error("Call should timeout or fail")
    }
}
