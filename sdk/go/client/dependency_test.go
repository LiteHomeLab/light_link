package client

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/nats-io/nats.go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDependencyType(t *testing.T) {
	dep := Dependency{
		ServiceName: "test-service",
		Methods:     []string{"method1", "method2"},
	}

	assert.Equal(t, "test-service", dep.ServiceName)
	assert.Equal(t, []string{"method1", "method2"}, dep.Methods)
}

func TestDependencyCheckResult(t *testing.T) {
	result := &DependencyCheckResult{
		ServiceName:      "test-service",
		ServiceFound:     true,
		AvailableMethods: []string{"method1"},
		MissingMethods:   []string{"method2"},
		AllSatisfied:     false,
	}

	assert.Equal(t, "test-service", result.ServiceName)
	assert.True(t, result.ServiceFound)
	assert.False(t, result.AllSatisfied)
}

func TestDependencyChecker_WaitForDependencies(t *testing.T) {
	// Skip if NATS not available
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create test NATS connection
	nc, err := nats.Connect(nats.DefaultURL)
	require.NoError(t, err)
	defer nc.Close()

	deps := []Dependency{
		{ServiceName: "test-service", Methods: []string{"method1"}},
	}

	checker := NewDependencyChecker(nc, deps)

	// Start waiting in background
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	doneCh := make(chan error, 1)
	go func() {
		doneCh <- checker.WaitForDependencies(ctx)
	}()

	// Give time for subscription to set up
	time.Sleep(200 * time.Millisecond)

	// Publish registration message
	metadata := &types.ServiceMetadata{
		Name:    "test-service",
		Version: "1.0.0",
		Methods: []types.MethodMetadata{
			{Name: "method1"},
		},
	}

	registerMsg := &types.RegisterMessage{
		Service:   "test-service",
		Version:   "1.0.0",
		Metadata:  *metadata,
		Timestamp: time.Now().Unix(),
	}

	data, err := json.Marshal(registerMsg)
	require.NoError(t, err)

	err = nc.Publish("$LL.register.test-service", data)
	require.NoError(t, err)

	// Should complete without error
	select {
	case err := <-doneCh:
		assert.NoError(t, err)
	case <-time.After(6 * time.Second):
		t.Fatal("Timeout waiting for dependencies")
	}
}
