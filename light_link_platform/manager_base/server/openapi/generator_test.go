package openapi

import (
	"encoding/json"
	"testing"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGenerateServiceOpenAPI(t *testing.T) {
	metadata := &types.ServiceMetadata{
		Name:        "test-service",
		Version:     "v1.0.0",
		Description: "Test service",
		Author:      "Test Author",
		Tags:        []string{"test"},
		Methods: []types.MethodMetadata{
			{
				Name:        "add",
				Description: "Add two numbers",
				Params: []types.ParameterMetadata{
					{Name: "a", Type: "number", Required: true, Description: "First number"},
					{Name: "b", Type: "number", Required: true, Description: "Second number"},
				},
				Returns: []types.ReturnMetadata{
					{Name: "sum", Type: "number", Description: "The sum"},
				},
			},
		},
	}

	spec := GenerateServiceOpenAPI(metadata)

	assert.Equal(t, "3.0.0", spec.OpenAPI)
	assert.Equal(t, "test-service", spec.Info.Title)
	assert.Equal(t, "v1.0.0", spec.Info.Version)

	// Check paths exist
	assert.Contains(t, spec.Paths, "/methods/add")
	addOp := spec.Paths["/methods/add"].Post
	assert.NotNil(t, addOp)
	assert.Equal(t, "add", addOp.OperationID)
	assert.Equal(t, "Add two numbers", addOp.Summary)

	// Verify JSON serialization works
	jsonBytes, err := json.MarshalIndent(spec, "", "  ")
	require.NoError(t, err)
	assert.Contains(t, string(jsonBytes), "openapi")
	assert.Contains(t, string(jsonBytes), "add")
}

func TestGenerateServiceOpenAPI_NoReturns(t *testing.T) {
	metadata := &types.ServiceMetadata{
		Name:        "test-service",
		Version:     "v1.0.0",
		Description: "Test service",
		Methods: []types.MethodMetadata{
			{
				Name:        "ping",
				Description: "Ping service",
				Params:      []types.ParameterMetadata{},
				Returns:     []types.ReturnMetadata{},
			},
		},
	}

	spec := GenerateServiceOpenAPI(metadata)

	assert.Contains(t, spec.Paths, "/methods/ping")
	pingOp := spec.Paths["/methods/ping"].Post
	assert.NotNil(t, pingOp)
	assert.Nil(t, pingOp.RequestBody)
}
