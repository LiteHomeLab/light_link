package service

import (
    "testing"
    "time"
)

func TestNewService(t *testing.T) {
    svc, err := NewService("test-service", "nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }
    defer svc.Stop()

    if svc.Name() != "test-service" {
        t.Errorf("Expected name 'test-service', got '%s'", svc.Name())
    }
}

func TestRegisterRPC(t *testing.T) {
    svc, err := NewService("test-service", "nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }
    defer svc.Stop()

    handler := func(args map[string]interface{}) (map[string]interface{}, error) {
        return map[string]interface{}{"result": "ok"}, nil
    }

    err = svc.RegisterRPC("testMethod", handler)
    if err != nil {
        t.Fatalf("RegisterRPC failed: %v", err)
    }

    // Verify handler is registered
    if !svc.HasRPC("testMethod") {
        t.Error("RPC method not registered")
    }
}

func TestStartStop(t *testing.T) {
    svc, err := NewService("test-service", "nats://localhost:4222", nil)
    if err != nil {
        t.Skip("Need running NATS server:", err)
    }

    err = svc.Start()
    if err != nil {
        t.Fatalf("Start failed: %v", err)
    }

    // Give service time to start
    time.Sleep(100 * time.Millisecond)

    err = svc.Stop()
    if err != nil {
        t.Fatalf("Stop failed: %v", err)
    }
}
