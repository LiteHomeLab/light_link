package types

import "fmt"

// ValidationError represents a parameter validation error
type ValidationError struct {
	ParameterName string      `json:"parameter_name"`
	ExpectedType  string      `json:"expected_type"`
	ActualType    string      `json:"actual_type"`
	ActualValue   interface{} `json:"actual_value,omitempty"`
	Message       string      `json:"message"`
}

// Error returns the error message
func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return fmt.Sprintf("parameter '%s': expected type %s, got %s",
		e.ParameterName, e.ExpectedType, e.ActualType)
}

// TypeInferenceResult represents the result of type inference
type TypeInferenceResult struct {
	TypeName  string `json:"type_name"`  // string, number, boolean, array, object
	Confidence int    `json:"confidence"` // 0-100
}

// ParameterTypeInference stores inferred parameter types
type ParameterTypeInference struct {
	Parameters map[string]TypeInferenceResult `json:"parameters"`
}

// ValidationErrorType is the type identifier for validation errors
const ValidationErrorType = "validation_error"
