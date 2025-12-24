package service

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
)

// TestRegisterMetadata tests metadata registration
func TestRegisterMetadata(t *testing.T) {
	// Create a NATS connection for testing
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("NATS not available:", err)
		return
	}
	defer nc.Close()

	// Subscribe to registration messages to capture them
	received := make(chan *types.RegisterMessage, 1)
	sub, err := nc.Subscribe("$LL.register.>", func(msg *nats.Msg) {
		var reg types.RegisterMessage
		if err := json.Unmarshal(msg.Data, &reg); err == nil {
			received <- &reg
		}
	})
	if err != nil {
		t.Fatal("Subscribe failed:", err)
	}
	defer sub.Unsubscribe()

	svc, err := NewService("test-service", nats.DefaultURL, nil)
	if err != nil {
		t.Fatal("NewService failed:", err)
	}
	defer svc.Stop()

	// Create metadata
	meta := &types.ServiceMetadata{
		Name:        "test-service",
		Version:     "v1.0.0",
		Description: "Test service",
		Author:      "Test Author",
		Tags:        []string{"test", "demo"},
	}

	// Register metadata
	if err := svc.RegisterMetadata(meta); err != nil {
		t.Fatal("RegisterMetadata failed:", err)
	}

	// Wait for message
	select {
	case reg := <-received:
		if reg.Service != "test-service" {
			t.Errorf("Expected service 'test-service', got '%s'", reg.Service)
		}
		if reg.Version != "v1.0.0" {
			t.Errorf("Expected version 'v1.0.0', got '%s'", reg.Version)
		}
	case <-time.After(2 * time.Second):
		t.Error("Did not receive registration message")
	}
}

// TestRegisterMethodWithMetadata tests method registration with metadata
func TestRegisterMethodWithMetadata(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("NATS not available:", err)
		return
	}
	defer nc.Close()

	svc, err := NewService("test-method-service", nats.DefaultURL, nil)
	if err != nil {
		t.Fatal("NewService failed:", err)
	}
	defer svc.Stop()

	// Create method metadata
	methodMeta := &types.MethodMetadata{
		Name:        "add",
		Description: "Add two numbers",
		Params: []types.ParameterMetadata{
			{Name: "a", Type: "number", Required: true, Description: "First number"},
			{Name: "b", Type: "number", Required: true, Description: "Second number"},
		},
		Returns: []types.ReturnMetadata{
			{Name: "result", Type: "number", Description: "Sum of the numbers"},
		},
		Example: &types.ExampleMetadata{
			Input:       map[string]any{"a": 10, "b": 20},
			Output:      map[string]any{"result": 30},
			Description: "10 + 20 = 30",
		},
		Tags: []string{"math", "basic"},
	}

	// Register method with metadata
	handler := func(args map[string]interface{}) (map[string]interface{}, error) {
		a := args["a"].(float64)
		b := args["b"].(float64)
		return map[string]interface{}{"result": a + b}, nil
	}

	if err := svc.RegisterMethodWithMetadata("add", handler, methodMeta); err != nil {
		t.Fatal("RegisterMethodWithMetadata failed:", err)
	}

	// Verify method is registered
	if !svc.HasRPC("add") {
		t.Error("Method 'add' was not registered")
	}

	// Verify metadata is stored
	meta, ok := svc.GetMethodMetadata("add")
	if !ok {
		t.Fatal("Method metadata not found")
	}

	if meta.Name != "add" {
		t.Errorf("Expected method name 'add', got '%s'", meta.Name)
	}

	if meta.Description != "Add two numbers" {
		t.Errorf("Expected description 'Add two numbers', got '%s'", meta.Description)
	}

	if len(meta.Params) != 2 {
		t.Errorf("Expected 2 params, got %d", len(meta.Params))
	}
}

// TestHeartbeat tests heartbeat sending
func TestHeartbeat(t *testing.T) {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		t.Skip("NATS not available:", err)
		return
	}
	defer nc.Close()

	// Subscribe to heartbeat messages
	received := make(chan *types.HeartbeatMessage, 5)
	sub, err := nc.Subscribe("$LL.heartbeat.>", func(msg *nats.Msg) {
		var hb types.HeartbeatMessage
		if err := json.Unmarshal(msg.Data, &hb); err == nil {
			received <- &hb
		}
	})
	if err != nil {
		t.Fatal("Subscribe failed:", err)
	}
	defer sub.Unsubscribe()

	svc, err := NewService("test-heartbeat", nats.DefaultURL, nil)
	if err != nil {
		t.Fatal("NewService failed:", err)
	}
	defer svc.Stop()

	// Start the service (which starts heartbeat)
	if err := svc.Start(); err != nil {
		t.Fatal("Start failed:", err)
	}

	// Wait for first heartbeat
	select {
	case hb := <-received:
		if hb.Service != "test-heartbeat" {
			t.Errorf("Expected service 'test-heartbeat', got '%s'", hb.Service)
		}
	case <-time.After(2 * time.Second):
		t.Error("Did not receive heartbeat message")
	}

	// Wait for second heartbeat (should come after interval)
	select {
	case <-received:
		// Good, received second heartbeat
	case <-time.After(35 * time.Second):
		t.Error("Did not receive second heartbeat")
	}
}

// TestBuildCurrentMetadata tests building metadata from registered methods
func TestBuildCurrentMetadata(t *testing.T) {
	svc, err := NewService("test-build-meta", nats.DefaultURL, nil)
	if err != nil {
		t.Skip("NATS not available:", err)
		return
	}
	defer svc.Stop()

	// Register some methods with metadata
	addMeta := &types.MethodMetadata{
		Name:        "add",
		Description: "Add two numbers",
		Params:      []types.ParameterMetadata{{Name: "a", Type: "number", Required: true}},
		Returns:     []types.ReturnMetadata{{Name: "result", Type: "number"}},
	}

	subMeta := &types.MethodMetadata{
		Name:        "subtract",
		Description: "Subtract two numbers",
		Params:      []types.ParameterMetadata{{Name: "a", Type: "number", Required: true}},
		Returns:     []types.ReturnMetadata{{Name: "result", Type: "number"}},
	}

	handler := func(args map[string]interface{}) (map[string]interface{}, error) {
		return map[string]interface{}{"result": 0}, nil
	}

	svc.RegisterMethodWithMetadata("add", handler, addMeta)
	svc.RegisterMethodWithMetadata("subtract", handler, subMeta)

	// Build metadata
	meta := svc.BuildCurrentMetadata(
		"test-build-meta",
		"v1.0.0",
		"Test service for metadata building",
		"Test Author",
		[]string{"test"},
	)

	if meta.Name != "test-build-meta" {
		t.Errorf("Expected name 'test-build-meta', got '%s'", meta.Name)
	}

	if len(meta.Methods) != 2 {
		t.Errorf("Expected 2 methods, got %d", len(meta.Methods))
	}
}

// TestGetSubjects tests subject generation
func TestGetSubjects(t *testing.T) {
	svc, err := NewService("test-subjects", nats.DefaultURL, nil)
	if err != nil {
		t.Fatal("NewService failed:", err)
	}
	defer svc.Stop()

	expectedHeartbeat := "$LL.heartbeat.test-subjects"
	if subj := svc.GetHeartbeatSubject(); subj != expectedHeartbeat {
		t.Errorf("Expected heartbeat subject '%s', got '%s'", expectedHeartbeat, subj)
	}

	expectedRegister := "$LL.register.test-subjects"
	if subj := svc.GetRegisterSubject(); subj != expectedRegister {
		t.Errorf("Expected register subject '%s', got '%s'", expectedRegister, subj)
	}
}
