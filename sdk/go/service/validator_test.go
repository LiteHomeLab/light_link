package service

import (
	"testing"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

func TestInferTypeString(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected string
	}{
		{"nil", nil, "null"},
		{"boolean true", true, "boolean"},
		{"boolean false", false, "boolean"},
		{"int", 42, "number"},
		{"int8", int8(8), "number"},
		{"int16", int16(16), "number"},
		{"int32", int32(32), "number"},
		{"int64", int64(64), "number"},
		{"uint", uint(42), "number"},
		{"uint8", uint8(8), "number"},
		{"uint16", uint16(16), "number"},
		{"uint32", uint32(32), "number"},
		{"uint64", uint64(64), "number"},
		{"float32", float32(3.14), "number"},
		{"float64", 3.14, "number"},
		{"string", "hello", "string"},
		{"array", []interface{}{1, 2, 3}, "array"},
		{"object", map[string]interface{}{"key": "value"}, "object"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := inferTypeString(tt.value)
			if result != tt.expected {
				t.Errorf("inferTypeString(%v) = %s, want %s", tt.value, result, tt.expected)
			}
		})
	}
}

func TestIsTypeCompatible(t *testing.T) {
	tests := []struct {
		name     string
		expected string
		actual   string
		want     bool
	}{
		{"same string", "string", "string", true},
		{"same number", "number", "number", true},
		{"same boolean", "boolean", "boolean", true},
		{"number compatible with integer", "number", "integer", true},
		{"number compatible with float", "number", "float", true},
		{"string not compatible with number", "string", "number", false},
		{"boolean not compatible with string", "boolean", "string", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isTypeCompatible(tt.expected, tt.actual)
			if result != tt.want {
				t.Errorf("isTypeCompatible(%s, %s) = %v, want %v", tt.expected, tt.actual, result, tt.want)
			}
		})
	}
}

func TestValidator_Validate(t *testing.T) {
	tests := []struct {
		name    string
		metadata *types.MethodMetadata
		args    map[string]interface{}
		wantErr bool
		errType string
	}{
		{
			name:     "nil metadata - no validation",
			metadata: nil,
			args:     map[string]interface{}{"a": "string"},
			wantErr:  false,
		},
		{
			name:     "empty params - no validation",
			metadata: &types.MethodMetadata{Params: []types.ParameterMetadata{}},
			args:     map[string]interface{}{"a": "string"},
			wantErr:  false,
		},
		{
			name: "valid string parameter",
			metadata: &types.MethodMetadata{
				Params: []types.ParameterMetadata{
					{Name: "name", Type: "string", Required: true},
				},
			},
			args:    map[string]interface{}{"name": "test"},
			wantErr: false,
		},
		{
			name: "valid number parameter",
			metadata: &types.MethodMetadata{
				Params: []types.ParameterMetadata{
					{Name: "count", Type: "number", Required: true},
				},
			},
			args:    map[string]interface{}{"count": 42},
			wantErr: false,
		},
		{
			name: "valid boolean parameter",
			metadata: &types.MethodMetadata{
				Params: []types.ParameterMetadata{
					{Name: "active", Type: "boolean", Required: true},
				},
			},
			args:    map[string]interface{}{"active": true},
			wantErr: false,
		},
		{
			name: "missing required parameter",
			metadata: &types.MethodMetadata{
				Params: []types.ParameterMetadata{
					{Name: "name", Type: "string", Required: true},
				},
			},
			args:    map[string]interface{}{},
			wantErr: true,
			errType: "*types.ValidationError",
		},
		{
			name: "type mismatch - string instead of number",
			metadata: &types.MethodMetadata{
				Params: []types.ParameterMetadata{
					{Name: "count", Type: "number", Required: true},
				},
			},
			args:    map[string]interface{}{"count": "not a number"},
			wantErr: true,
			errType: "*types.ValidationError",
		},
		{
			name: "type mismatch - number instead of string",
			metadata: &types.MethodMetadata{
				Params: []types.ParameterMetadata{
					{Name: "name", Type: "string", Required: true},
				},
			},
			args:    map[string]interface{}{"name": 123},
			wantErr: true,
			errType: "*types.ValidationError",
		},
		{
			name: "optional parameter not provided",
			metadata: &types.MethodMetadata{
				Params: []types.ParameterMetadata{
					{Name: "name", Type: "string", Required: true},
					{Name: "optional", Type: "string", Required: false},
				},
			},
			args:    map[string]interface{}{"name": "test"},
			wantErr: false,
		},
		{
			name: "multiple valid parameters",
			metadata: &types.MethodMetadata{
				Params: []types.ParameterMetadata{
					{Name: "a", Type: "number", Required: true},
					{Name: "b", Type: "number", Required: true},
				},
			},
			args:    map[string]interface{}{"a": 1, "b": 2},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := NewValidator(tt.metadata)
			err := v.Validate(tt.args)

			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr && tt.errType != "" {
				// Check error type
				if _, ok := err.(*types.ValidationError); !ok {
					t.Errorf("Validate() error type = %T, want %s", err, tt.errType)
				}
			}
		})
	}
}

func TestInferTypesFromExample(t *testing.T) {
	example := &types.ExampleMetadata{
		Input: map[string]interface{}{
			"name":  "test",
			"count": 42,
			"active": true,
			"items": []interface{}{1, 2, 3},
			"meta":  map[string]interface{}{"key": "value"},
		},
	}

	result := InferTypesFromExample(example)

	if result == nil {
		t.Fatal("InferTypesFromExample() returned nil")
	}

	expectedTypes := map[string]string{
		"name":  "string",
		"count": "number",
		"active": "boolean",
		"items": "array",
		"meta":  "object",
	}

	for name, expectedType := range expectedTypes {
		inferred, ok := result.Parameters[name]
		if !ok {
			t.Errorf("InferTypesFromExample() missing parameter %s", name)
			continue
		}
		if inferred.TypeName != expectedType {
			t.Errorf("InferTypesFromExample() %s type = %s, want %s", name, inferred.TypeName, expectedType)
		}
		if inferred.Confidence != 80 {
			t.Errorf("InferTypesFromExample() %s confidence = %d, want 80", name, inferred.Confidence)
		}
	}
}

func TestInferTypesFromExample_Nil(t *testing.T) {
	result := InferTypesFromExample(nil)
	if result != nil {
		t.Errorf("InferTypesFromExample(nil) = %v, want nil", result)
	}

	result = InferTypesFromExample(&types.ExampleMetadata{})
	if result != nil {
		t.Errorf("InferTypesFromExample(empty) = %v, want nil", result)
	}
}
