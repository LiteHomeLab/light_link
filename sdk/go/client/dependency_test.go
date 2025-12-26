package client

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
