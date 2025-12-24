package service

import (
	"fmt"
	"reflect"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

// Validator validates RPC parameters
type Validator struct {
	methodMeta *types.MethodMetadata
}

// NewValidator creates a new parameter validator
func NewValidator(methodMeta *types.MethodMetadata) *Validator {
	return &Validator{methodMeta: methodMeta}
}

// Validate validates args against method metadata
func (v *Validator) Validate(args map[string]interface{}) error {
	if v.methodMeta == nil || len(v.methodMeta.Params) == 0 {
		return nil // No metadata to validate against
	}

	for _, paramMeta := range v.methodMeta.Params {
		value, exists := args[paramMeta.Name]

		// Check required
		if paramMeta.Required && !exists {
			return &types.ValidationError{
				ParameterName: paramMeta.Name,
				ExpectedType:  paramMeta.Type,
				ActualType:    "missing",
				Message:       fmt.Sprintf("required parameter '%s' is missing", paramMeta.Name),
			}
		}

		if !exists {
			continue // Optional parameter not provided
		}

		// Check type
		actualType := inferTypeString(value)
		if !isTypeCompatible(paramMeta.Type, actualType) {
			return &types.ValidationError{
				ParameterName: paramMeta.Name,
				ExpectedType:  paramMeta.Type,
				ActualType:    actualType,
				ActualValue:   value,
				Message:       fmt.Sprintf("parameter '%s': expected type %s, got %s",
					paramMeta.Name, paramMeta.Type, actualType),
			}
		}
	}

	return nil
}

// inferTypeString infers the type string from a value
func inferTypeString(value interface{}) string {
	if value == nil {
		return "null"
	}

	switch v := value.(type) {
	case bool:
		return "boolean"
	case int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64:
		return "number"
	case float32, float64:
		return "number"
	case string:
		return "string"
	case []interface{}:
		return "array"
	case map[string]interface{}:
		return "object"
	default:
		// Use reflection for other types
		t := reflect.TypeOf(v)
		switch t.Kind() {
		case reflect.Bool:
			return "boolean"
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return "number"
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return "number"
		case reflect.Float32, reflect.Float64:
			return "number"
		case reflect.String:
			return "string"
		case reflect.Slice, reflect.Array:
			return "array"
		case reflect.Map, reflect.Struct:
			return "object"
		default:
			return "unknown"
		}
	}
}

// isTypeCompatible checks if actual type is compatible with expected type
func isTypeCompatible(expected, actual string) bool {
	// Direct match
	if expected == actual {
		return true
	}

	// Number types: int and uint are compatible with number
	if expected == "number" && (actual == "integer" || actual == "float") {
		return true
	}

	return false
}

// InferTypesFromExample infers parameter types from example data
func InferTypesFromExample(example *types.ExampleMetadata) *types.ParameterTypeInference {
	if example == nil || example.Input == nil {
		return nil
	}

	result := &types.ParameterTypeInference{
		Parameters: make(map[string]types.TypeInferenceResult),
	}

	for name, value := range example.Input {
		typeName := inferTypeString(value)
		result.Parameters[name] = types.TypeInferenceResult{
			TypeName:  typeName,
			Confidence: 80, // Medium confidence from example
		}
	}

	return result
}
