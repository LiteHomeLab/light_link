package openapi

import (
	"encoding/json"

	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

// OpenAPI represents OpenAPI 3.0 specification
type OpenAPI struct {
	OpenAPI string `json:"openapi"`
	Info    Info   `json:"info"`
	Paths   Paths  `json:"paths"`
}

// Info contains metadata about the API
type Info struct {
	Title       string   `json:"title"`
	Version     string   `json:"version"`
	Description string   `json:"description,omitempty"`
}

// Paths maps path to path item
type Paths map[string]PathItem

// PathItem describes operations on a single path
type PathItem struct {
	Post *Operation `json:"post,omitempty"`
}

// Operation describes a single API operation
type Operation struct {
	OperationID string              `json:"operationId"`
	Summary     string              `json:"summary"`
	Description string              `json:"description,omitempty"`
	Tags        []string            `json:"tags,omitempty"`
	RequestBody *RequestBody        `json:"requestBody,omitempty"`
	Responses   map[string]Response `json:"responses"`
}

// RequestBody describes request body
type RequestBody struct {
	Content map[string]MediaType `json:"content"`
}

// MediaType represents a media type
type MediaType struct {
	Schema *Schema `json:"schema,omitempty"`
}

// Schema represents JSON Schema
type Schema struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
	Required   []string            `json:"required,omitempty"`
}

// Property is a schema property
type Property struct {
	Type        string `json:"type"`
	Description string `json:"description,omitempty"`
}

// Response represents an API response
type Response struct {
	Description string              `json:"description"`
	Content     map[string]MediaType `json:"content,omitempty"`
}

// GenerateServiceOpenAPI generates OpenAPI spec for a service
func GenerateServiceOpenAPI(metadata *types.ServiceMetadata) *OpenAPI {
	spec := &OpenAPI{
		OpenAPI: "3.0.0",
		Info: Info{
			Title:       metadata.Name,
			Version:     metadata.Version,
			Description: metadata.Description,
		},
		Paths: make(Paths),
	}

	// Generate path for each method
	for _, method := range metadata.Methods {
		path := "/methods/" + method.Name
		spec.Paths[path] = PathItem{
			Post: generateOperation(method),
		}
	}

	return spec
}

// generateOperation creates an operation from method metadata
func generateOperation(method types.MethodMetadata) *Operation {
	op := &Operation{
		OperationID: method.Name,
		Summary:     method.Description,
		Tags:        method.Tags,
		Responses: map[string]Response{
			"200": {
				Description: "Success",
				Content: map[string]MediaType{
					"application/json": {
						Schema: generateResponseSchema(method.Returns),
					},
				},
			},
		},
	}

	// Add request body if has parameters
	if len(method.Params) > 0 {
		op.RequestBody = &RequestBody{
			Content: map[string]MediaType{
				"application/json": {
					Schema: generateRequestSchema(method.Params),
				},
			},
		}
	}

	return op
}

// generateRequestSchema creates schema for request parameters
func generateRequestSchema(params []types.ParameterMetadata) *Schema {
	if len(params) == 0 {
		return nil
	}

	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]Property),
		Required:   make([]string, 0),
	}

	for _, param := range params {
		schema.Properties[param.Name] = Property{
			Type:        param.Type,
			Description: param.Description,
		}
		if param.Required {
			schema.Required = append(schema.Required, param.Name)
		}
	}

	return schema
}

// generateResponseSchema creates schema for response
func generateResponseSchema(returns []types.ReturnMetadata) *Schema {
	if len(returns) == 0 {
		return &Schema{Type: "object"}
	}

	schema := &Schema{
		Type:       "object",
		Properties: make(map[string]Property),
	}

	for _, ret := range returns {
		schema.Properties[ret.Name] = Property{
			Type:        ret.Type,
			Description: ret.Description,
		}
	}

	return schema
}

// ToJSON converts OpenAPI spec to JSON bytes
func (o *OpenAPI) ToJSON() ([]byte, error) {
	return json.MarshalIndent(o, "", "  ")
}

// ToYAML converts OpenAPI spec to YAML
func (o *OpenAPI) ToYAML() ([]byte, error) {
	// For now, return JSON only
	// Can add yaml.v3 support later
	return o.ToJSON()
}
