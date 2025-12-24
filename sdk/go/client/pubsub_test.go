package client

import (
    "testing"
    "time"
)

func TestPublishSubscribe(t *testing.T) {
    config := &TLSConfig{
        CaFile:     "../../../deploy/nats/tls/ca.crt",
        CertFile:   "../../../deploy/nats/tls/demo-service.crt",
        KeyFile:    "../../../deploy/nats/tls/demo-service.key",
        ServerName: "nats-server",
    }
    subClient, err := NewClient("nats://172.18.200.47:4222", config)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }
    defer subClient.Close()

    pubClient, err := NewClient("nats://172.18.200.47:4222", config)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }
    defer pubClient.Close()

    received := make(chan map[string]interface{}, 1)

    // Subscribe
    sub, err := subClient.Subscribe("test.subject", func(data map[string]interface{}) {
        received <- data
    })
    if err != nil {
        t.Fatalf("Subscribe failed: %v", err)
    }
    defer sub.Unsubscribe()

    // Wait a bit for subscription to be fully established
    time.Sleep(100 * time.Millisecond)

    // Publish
    err = pubClient.Publish("test.subject", map[string]interface{}{"msg": "hello"})
    if err != nil {
        t.Fatalf("Publish failed: %v", err)
    }

    // Wait for receive
    select {
    case data := <-received:
        if data["msg"] != "hello" {
            t.Errorf("Expected 'hello', got '%v'", data["msg"])
        }
    case <-time.After(2 * time.Second):
        t.Error("Timeout waiting for message")
    }
}
