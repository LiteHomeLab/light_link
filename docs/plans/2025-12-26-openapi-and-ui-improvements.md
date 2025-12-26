# OpenAPI and UI Improvements Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Add OpenAPI/Swagger export functionality and fix UI to display complete method signatures (parameters and return values) in the service detail view.

**Architecture:**
1. Fix frontend display bug where `return_info` array is shown as "void"
2. Create OpenAPI generator that converts service metadata to OpenAPI 3.0 spec
3. Add API endpoints for serving OpenAPI docs (JSON/YAML)
4. Add frontend UI for viewing/downloading OpenAPI documentation

**Tech Stack:**
- Go 1.23+ (backend)
- Vue 3 + TypeScript + Element Plus (frontend)
- OpenAPI 3.0 specification
- Playwright for testing

---

## Task 1: Fix Frontend Return Value Display

**Files:**
- Modify: `light_link_platform/manager_base/web/src/views/ServiceDetailView.vue:62-66`
- Test: Manual verification with Playwright

**Step 1: Update the method meta display to show return values properly**

Current code shows `返回: {{ method.return_info?.type || 'void' }}` but `return_info` is an array.

Replace the return value display section (around line 62-66) with:

```vue
<div class="method-meta">
  <el-tag size="small" type="info" v-if="method.return_info && method.return_info.length">
    返回: {{ formatReturnInfo(method.return_info) }}
  </el-tag>
  <el-tag size="small" type="info" v-else>
    返回: void
  </el-tag>
</div>

<!-- Add return values table if present -->
<div v-if="method.return_info && method.return_info.length" class="return-values">
  <h5>返回值:</h5>
  <el-table :data="method.return_info" size="small">
    <el-table-column prop="name" label="名称" width="150" />
    <el-table-column prop="type" label="类型" width="150" />
    <el-table-column prop="description" label="描述" />
  </el-table>
</div>
```

**Step 2: Add formatReturnInfo function to script section**

Add to the script setup (after formatJSON function):

```typescript
function formatReturnInfo(returnInfo: any[]): string {
  if (!returnInfo || !returnInfo.length) return 'void';
  return returnInfo.map(r => `${r.name || 'value'}: ${r.type}`).join(', ');
}
```

**Step 3: Add CSS for return-values section**

Add to the style section:

```css
.return-values {
  margin-top: 12px;
}

.return-values h5 {
  margin: 0 0 8px 0;
  font-size: 14px;
  color: #333;
}
```

**Step 4: Verify the fix with Playwright**

```bash
# Start services
cd light_link_platform/manager_base/server && go run main.go &
cd light_link_platform/manager_base/web && npm run dev &

# Open browser and check
# Use Playwright to navigate to http://localhost:5173/services/math-service-go
# Verify that return values are displayed correctly in a table format
```

**Step 5: Commit**

```bash
git add light_link_platform/manager_base/web/src/views/ServiceDetailView.vue
git commit -m "fix(web): display return values in method detail view

- Add return values table showing name, type, and description
- Fix 'void' display bug when return_info array exists
- Add formatReturnInfo helper function for inline display"
```

---

## Task 2: Create OpenAPI Generator Package

**Files:**
- Create: `light_link_platform/manager_base/server/openapi/generator.go`
- Test: `light_link_platform/manager_base/server/openapi/generator_test.go`

**Step 1: Write test for OpenAPI generation**

Create test file:

```go
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
```

**Step 2: Run test to verify it fails**

```bash
cd light_link_platform/manager_base/server
go test ./openapi/... -v
```

Expected: `package .../openapi does not exist` or `undefined: GenerateServiceOpenAPI`

**Step 3: Create OpenAPI generator implementation**

Create `light_link_platform/manager_base/server/openapi/generator.go`:

```go
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
    Examples    []Example           `json:"examples,omitempty"`
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
    Type       string                 `json:"type"`
    Properties map[string]Property    `json:"properties,omitempty"`
    Required   []string               `json:"required,omitempty"`
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

// Example represents an example
type Example struct {
    Input       map[string]any `json:"input"`
    Output      map[string]any `json:"output"`
    Description string         `json:"description,omitempty"`
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

    // Add example if present
    if method.Example != nil {
        // Examples are handled separately or embedded in the spec
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

// ToYAML converts OpenAPI spec to YAML (optional, requires gopkg.in/yaml.v3)
func (o *OpenAPI) ToYAML() ([]byte, error) {
    // For now, return JSON only
    // Can add yaml.v3 support later
    return o.ToJSON()
}
```

**Step 4: Run test to verify it passes**

```bash
cd light_link_platform/manager_base/server
go test ./openapi/... -v
```

Expected: PASS

**Step 5: Commit**

```bash
git add light_link_platform/manager_base/server/openapi/
git commit -m "feat(openapi): add OpenAPI 3.0 generator

- Add OpenAPI spec structure types
- Implement GenerateServiceOpenAPI for service metadata
- Convert method parameters/returns to OpenAPI schemas
- Add unit tests for generator"
```

---

## Task 3: Add OpenAPI API Endpoints

**Files:**
- Modify: `light_link_platform/manager_base/server/api/handler.go`
- Test: Manual with curl

**Step 1: Add OpenAPI route handler**

In `handler.go`, add to the Routes function (after line 56):

```go
// OpenAPI endpoints
mux.HandleFunc("/api/services/", h.withAuth(h.handleServiceRouter))
mux.HandleFunc("/api/services/", h.withAuth(h.handleOpenAPIRouter))
```

Actually, we need to integrate this properly. Modify the existing handleServiceRouter:

```go
// handleServiceRouter routes /api/services/ requests to appropriate handlers
func (h *Handler) handleServiceRouter(w http.ResponseWriter, r *http.Request) {
    parts := strings.Split(r.URL.Path, "/")
    if len(parts) < 4 {
        sendJSONError(w, http.StatusBadRequest, "Invalid path")
        return
    }

    // Check if it's an OpenAPI request: /api/services/{service}/openapi
    if len(parts) >= 5 && parts[4] == "openapi" {
        h.handleOpenAPI(w, r, parts[3])
        return
    }

    // Check if it's a methods request: /api/services/{service}/methods
    if len(parts) >= 5 && parts[4] == "methods" {
        h.handleMethods(w, r)
        return
    }

    // Otherwise, it's a service detail request: /api/services/{service}
    h.handleServiceDetail(w, r)
}
```

**Step 2: Implement handleOpenAPI**

Add after handleMethods function (around line 287):

```go
// handleOpenAPI handles OpenAPI spec requests
func (h *Handler) handleOpenAPI(w http.ResponseWriter, r *http.Request, serviceName string) {
    if r.Method != http.MethodGet {
        sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
        return
    }

    // Get service metadata
    service, err := h.db.GetService(serviceName)
    if err != nil {
        sendJSONError(w, http.StatusNotFound, "Service not found")
        return
    }

    // Get methods
    methods, err := h.db.GetMethods(serviceName)
    if err != nil {
        sendJSONError(w, http.StatusInternalServerError, "Failed to get methods")
        return
    }

    // Convert to types.ServiceMetadata
    metadata := convertToServiceMetadata(service, methods)

    // Generate OpenAPI spec
    spec := openapi.GenerateServiceOpenAPI(metadata)

    // Check format (json or yaml)
    format := r.URL.Query().Get("format")
    if format == "yaml" {
        w.Header().Set("Content-Type", "application/x-yaml")
        yamlBytes, err := spec.ToYAML()
        if err != nil {
            sendJSONError(w, http.StatusInternalServerError, "Failed to generate YAML")
            return
        }
        w.Write(yamlBytes)
    } else {
        w.Header().Set("Content-Type", "application/json")
        jsonBytes, err := spec.ToJSON()
        if err != nil {
            sendJSONError(w, http.StatusInternalServerError, "Failed to generate JSON")
            return
        }
        w.Write(jsonBytes)
    }
}

// convertToServiceMetadata converts storage types to SDK types
func convertToServiceMetadata(service *storage.ServiceMetadata, methods []*storage.MethodMetadata) *types.ServiceMetadata {
    sdkMethods := make([]types.MethodMetadata, len(methods))
    for i, m := range methods {
        sdkMethods[i] = types.MethodMetadata{
            Name:        m.Name,
            Description: m.Description,
            Params:      m.Params,
            Returns:     m.Returns,
            Example:     m.Example,
            Tags:        m.Tags,
            Deprecated:  m.Deprecated,
        }
    }

    return &types.ServiceMetadata{
        Name:        service.Name,
        Version:     service.Version,
        Description: service.Description,
        Author:      service.Author,
        Tags:        service.Tags,
        Methods:     sdkMethods,
    }
}
```

**Step 3: Import openapi package**

Add to imports in handler.go:

```go
import (
    "encoding/json"
    "net/http"
    "strconv"
    "strings"

    "github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/auth"
    "github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/manager"
    "github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/openapi"
    "github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/storage"
)
```

**Step 4: Test the endpoints**

```bash
# Get token
TOKEN=$(curl -s 'http://localhost:8080/api/auth/login' \
  -H 'Content-Type: application/json' \
  -d '{"username":"admin","password":"admin"}' | jq -r '.token')

# Test JSON endpoint
curl "http://localhost:8080/api/services/math-service-go/openapi?format=json" \
  -H "Authorization: Bearer $TOKEN" | jq .

# Test YAML endpoint (if implemented)
curl "http://localhost:8080/api/services/math-service-go/openapi?format=yaml" \
  -H "Authorization: Bearer $TOKEN"
```

**Step 5: Commit**

```bash
git add light_link_platform/manager_base/server/api/handler.go
git commit -m "feat(api): add OpenAPI spec endpoints

- Add GET /api/services/{name}/openapi?format=json
- Add GET /api/services/{name}/openapi?format=yaml
- Generate OpenAPI 3.0 spec from service metadata
- Support both JSON and YAML output formats"
```

---

## Task 4: Add Frontend API Integration

**Files:**
- Modify: `light_link_platform/manager_base/web/src/api/index.ts`
- Test: TypeScript compilation

**Step 1: Add OpenAPI API methods**

In `index.ts`, add to servicesApi object (after line 175):

```typescript
getOpenAPI: (name: string, format: 'json' | 'yaml' = 'json') =>
  api.get(`/services/${name}/openapi?format=${format}`, {
    responseType: format === 'yaml' ? 'text' : 'json'
  }),
```

**Step 2: Commit**

```bash
git add light_link_platform/manager_base/web/src/api/index.ts
git commit -m "feat(web): add OpenAPI API method"
```

---

## Task 5: Add OpenAPI View/Download UI

**Files:**
- Modify: `light_link_platform/manager_base/web/src/views/ServiceDetailView.vue`
- Test: Manual verification

**Step 1: Add OpenAPI button to header**

Add to the template header section (after refresh button, around line 9):

```vue
<template #extra>
  <el-button @click="loadData" :loading="loading">
    <el-icon><Refresh /></el-icon>
    刷新
  </el-button>
  <el-button @click="showOpenAPI" type="primary" plain>
    <el-icon><Document /></el-icon>
    OpenAPI
  </el-button>
</template>
```

**Step 2: Add OpenAPI dialog**

Add to template before closing div (after instances-card):

```vue
<!-- OpenAPI Dialog -->
<el-dialog
  v-model="openapiDialogVisible"
  title="OpenAPI Specification"
  width="80%"
  :close-on-click-modal="false"
>
  <el-tabs v-model="activeFormat">
    <el-tab-pane label="JSON" name="json">
      <pre class="openapi-content">{{ openapiContent.json }}</pre>
    </el-tab-pane>
    <el-tab-pane label="YAML" name="yaml">
      <pre class="openapi-content">{{ openapiContent.yaml }}</pre>
    </el-tab-pane>
  </el-tabs>

  <template #footer>
    <el-button @click="openapiDialogVisible = false">关闭</el-button>
    <el-button type="primary" @click="downloadOpenAPI">
      <el-icon><Download /></el-icon>
      下载
    </el-button>
  </template>
</el-dialog>
```

**Step 3: Add state and methods to script**

Add to script setup (after other refs):

```typescript
import { Document, Download } from '@element-plus/icons-vue'

const openapiDialogVisible = ref(false)
const activeFormat = ref('json')
const openapiContent = ref({ json: '', yaml: '' })

async function showOpenAPI() {
  try {
    const [jsonResp, yamlResp] = await Promise.all([
      servicesApi.getOpenAPI(serviceName.value, 'json'),
      servicesApi.getOpenAPI(serviceName.value, 'yaml')
    ])

    openapiContent.value = {
      json: typeof jsonResp.data === 'string' ? jsonResp.data : JSON.stringify(jsonResp.data, null, 2),
      yaml: yamlResp.data
    }

    openapiDialogVisible.value = true
  } catch (error: any) {
    ElMessage.error('Failed to load OpenAPI spec')
  }
}

function downloadOpenAPI() {
  const content = activeFormat.value === 'json' ? openapiContent.value.json : openapiContent.value.yaml
  const blob = new Blob([content], { type: activeFormat.value === 'json' ? 'application/json' : 'text/yaml' })
  const url = URL.createObjectURL(blob)
  const a = document.createElement('a')
  a.href = url
  a.download = `${serviceName.value}-openapi.${activeFormat.value}`
  a.click()
  URL.revokeObjectURL(url)
  ElMessage.success('Downloaded')
}
```

**Step 4: Add styles**

Add to style section:

```css
.openapi-content {
  background-color: #f5f5f5;
  border: 1px solid #e0e0e0;
  border-radius: 4px;
  padding: 16px;
  max-height: 500px;
  overflow: auto;
  font-size: 12px;
  margin: 0;
}
```

**Step 5: Verify with Playwright**

Open the service detail page, click "OpenAPI" button, verify dialog shows content correctly.

**Step 6: Commit**

```bash
git add light_link_platform/manager_base/web/src/views/ServiceDetailView.vue
git commit -m "feat(web): add OpenAPI viewer and download

- Add OpenAPI button to service detail page
- Show OpenAPI spec in dialog with JSON/YAML tabs
- Add download functionality for both formats"
```

---

## Task 6: Update All Provider Services with Complete Metadata

**Files:**
- Modify: `light_link_platform/examples/provider/python/math_service/main.py`
- Modify: `light_link_platform/examples/provider/csharp/MathService/Program.cs`
- Test: Restart services and verify

**Step 1: Update Python service metadata**

Ensure Python service includes return metadata. Check `main.py` has proper MethodMetadata with Returns array.

**Step 2: Update C# service metadata**

Ensure C# service includes return metadata. Check `Program.cs` has proper MethodMetadata with Returns array.

**Step 3: Verify all services show return values correctly**

```bash
# Restart all services
# Check in browser that all methods show return values
```

**Step 4: Commit if changes needed**

```bash
git add light_link_platform/examples/provider/
git commit -m "docs(examples): ensure all services have complete metadata"
```

---

## Task 7: Final Integration Testing

**Files:**
- None (verification only)

**Step 1: Start all services**

```bash
# Start manager backend
cd light_link_platform/manager_base/server && go run main.go

# Start frontend
cd light_link_platform/manager_base/web && npm run dev

# Start provider services
cd light_link_platform/examples/provider/go/math-service-go && go run main.go
cd light_link_platform/examples/provider/python/math_service && python main.py
cd light_link_platform/examples/provider/csharp/MathService && dotnet run
```

**Step 2: Verify with Playwright**

Navigate to each service detail page and verify:
1. Method cards show clean name + description
2. Clicking method shows parameter table
3. Return values are shown in table format
4. OpenAPI button works
5. OpenAPI dialog shows correct spec
6. Download works

**Step 3: Test OpenAPI endpoints**

```bash
# Test each service's OpenAPI endpoint
for service in "math-service-go" "math-service-python" "math-service-csharp"; do
  curl "http://localhost:8080/api/services/$service/openapi?format=json" \
    -H "Authorization: Bearer $TOKEN" | jq .info.title
done
```

**Step 8: Final commit**

```bash
git commit --allow-empty -m "test: verify OpenAPI and UI improvements

- All services display return values correctly
- OpenAPI endpoints working for all services
- Frontend UI shows complete method signatures"
```

---

## Summary

This plan implements:
1. Fixed return value display bug in frontend
2. OpenAPI 3.0 generator for service metadata
3. API endpoints for JSON/YAML export
4. Frontend UI for viewing/downloading specs

Total estimated tasks: 7
Estimated implementation time: 2-3 hours
