package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/auth"
	"github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/manager"
	"github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/openapi"
	"github.com/LiteHomeLab/light_link/light_link_platform/manager_base/server/storage"
	"github.com/LiteHomeLab/light_link/sdk/go/types"
)

// Handler handles API requests
type Handler struct {
	db         *storage.Database
	manager    *manager.Manager
	auth       *auth.AuthMiddleware
	controller *manager.Controller
}

// NewHandler creates a new API handler
func NewHandler(db *storage.Database, mgr *manager.Manager, auth *auth.AuthMiddleware) *Handler {
	ctrl := manager.NewController(mgr)
	return &Handler{
		db:         db,
		manager:    mgr,
		auth:       auth,
		controller: ctrl,
	}
}

// Routes returns the HTTP handler with all routes registered
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("/api/auth/login", h.handleLogin)

	// Service endpoints
	mux.HandleFunc("/api/services", h.withAuth(h.handleServices))
	mux.HandleFunc("/api/services/", h.withAuth(h.handleServiceRouter))

	// Status endpoints
	mux.HandleFunc("/api/status", h.withAuth(h.handleStatus))
	mux.HandleFunc("/api/status/", h.withAuth(h.handleServiceStatus))

	// Event endpoints
	mux.HandleFunc("/api/events", h.withAuth(h.handleEvents))

	// Call endpoint
	mux.HandleFunc("/api/call", h.withAuth(h.handleCall))

	// Instance endpoints
	mux.HandleFunc("/api/instances", h.withAuth(h.handleInstances))
	mux.HandleFunc("/api/instances/", h.withAuth(h.handleInstanceRouter))

	// WebSocket endpoint (auth handled separately)
	mux.HandleFunc("/api/ws", h.handleWebSocket)

	return h.auth.Middleware()(mux)
}

// withAuth wraps a handler with auth middleware
func (h *Handler) withAuth(fn http.HandlerFunc) http.HandlerFunc {
	return fn
}

// sendJSON sends a JSON response
func sendJSON(w http.ResponseWriter, data interface{}) error {
	w.Header().Set("Content-Type", "application/json")
	return json.NewEncoder(w).Encode(data)
}

// sendJSONError sends a JSON error response
func sendJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// handleLogin handles login requests
func (h *Handler) handleLogin(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	var req struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Validate user
	user, err := h.db.ValidateUser(req.Username, req.Password)
	if err != nil {
		sendJSONError(w, http.StatusUnauthorized, "Invalid username or password")
		return
	}

	// Generate token
	token, err := h.auth.GenerateToken(user.Username, user.Role)
	if err != nil {
		sendJSONError(w, http.StatusInternalServerError, "Failed to generate token")
		return
	}

	json.NewEncoder(w).Encode(map[string]string{"token": token, "role": user.Role})
}

// handleServices handles service list requests
func (h *Handler) handleServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	services, err := h.db.ListServices()
	if err != nil {
		sendJSONError(w, http.StatusInternalServerError, "Failed to get services")
		return
	}

	sendJSON(w, services)
}

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

// handleServiceDetail handles service detail requests
func (h *Handler) handleServiceDetail(w http.ResponseWriter, r *http.Request) {
	// Extract service name from path
	// /api/services/demo-service -> demo-service
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		sendJSONError(w, http.StatusBadRequest, "Invalid service name")
		return
	}
	serviceName := parts[3]

	switch r.Method {
	case http.MethodGet:
		h.handleGetService(w, r, serviceName)
	case http.MethodDelete:
		h.handleDeleteService(w, r, serviceName)
	default:
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleGetService retrieves a single service
func (h *Handler) handleGetService(w http.ResponseWriter, r *http.Request, serviceName string) {
	service, err := h.db.GetService(serviceName)
	if err != nil {
		sendJSONError(w, http.StatusNotFound, "Service not found")
		return
	}

	sendJSON(w, service)
}

// handleDeleteService deletes a service (only if offline)
func (h *Handler) handleDeleteService(w http.ResponseWriter, r *http.Request, serviceName string) {
	// Check if service is offline before allowing deletion
	status, err := h.db.GetServiceStatus(serviceName)
	if err != nil {
		sendJSONError(w, http.StatusNotFound, "Service not found")
		return
	}

	if status.Online {
		sendJSONError(w, http.StatusForbidden, "Cannot delete online service")
		return
	}

	// Delete service and all related data
	if err := h.db.DeleteServiceCascade(serviceName); err != nil {
		sendJSONError(w, http.StatusInternalServerError, "Failed to delete service")
		return
	}

	sendJSON(w, map[string]string{"message": "Service deleted successfully"})
}

// handleStatus handles status list requests
func (h *Handler) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	statuses, err := h.db.ListServiceStatus()
	if err != nil {
		sendJSONError(w, http.StatusInternalServerError, "Failed to get status")
		return
	}

	sendJSON(w, statuses)
}

// handleServiceStatus handles service status requests
func (h *Handler) handleServiceStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract service name from path
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		sendJSONError(w, http.StatusBadRequest, "Invalid service name")
		return
	}
	serviceName := parts[3]

	status, err := h.db.GetServiceStatus(serviceName)
	if err != nil {
		sendJSONError(w, http.StatusNotFound, "Service status not found")
		return
	}

	sendJSON(w, status)
}

// handleMethods handles method requests
func (h *Handler) handleMethods(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract service name from path
	// /api/services/demo-service/methods -> demo-service
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 5 || parts[4] != "methods" {
		// Try getting specific method
		if len(parts) >= 6 {
			serviceName := parts[3]
			methodName := parts[5]
			h.getMethodDetail(w, r, serviceName, methodName)
			return
		}
		sendJSONError(w, http.StatusBadRequest, "Invalid path")
		return
	}
	serviceName := parts[3]

	methods, err := h.db.GetMethods(serviceName)
	if err != nil {
		sendJSONError(w, http.StatusInternalServerError, "Failed to get methods")
		return
	}

	sendJSON(w, methods)
}

// getMethodDetail returns a specific method
func (h *Handler) getMethodDetail(w http.ResponseWriter, r *http.Request, serviceName, methodName string) {
	method, err := h.db.GetMethod(serviceName, methodName)
	if err != nil {
		sendJSONError(w, http.StatusNotFound, "Method not found")
		return
	}

	sendJSON(w, method)
}

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

// handleEvents handles event requests
func (h *Handler) handleEvents(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Parse query parameters
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	limit := 100
	offset := 0

	if limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}
	if offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	events, err := h.db.ListEvents(limit, offset)
	if err != nil {
		sendJSONError(w, http.StatusInternalServerError, "Failed to get events")
		return
	}

	sendJSON(w, events)
}

// handleCall handles RPC call requests
func (h *Handler) handleCall(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if !auth.IsAdmin(r) {
		sendJSONError(w, http.StatusForbidden, "Admin access required")
		return
	}

	var req struct {
		Service string                 `json:"service"`
		Method  string                 `json:"method"`
		Params  map[string]interface{} `json:"params"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		sendJSONError(w, http.StatusBadRequest, "Invalid request")
		return
	}

	// Check if service is online
	status, err := h.db.GetServiceStatus(req.Service)
	if err != nil {
		sendJSONError(w, http.StatusServiceUnavailable, "Service not found")
		return
	}
	if !status.Online {
		sendJSONError(w, http.StatusServiceUnavailable, "Service is offline")
		return
	}

	// Make the RPC call
	result, err := h.manager.CallServiceMethod(req.Service, req.Method, req.Params)
	if err != nil {
		sendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	// Return the result
	json.NewEncoder(w).Encode(result)
}

// handleWebSocket handles WebSocket connections
func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket upgrade
	// This will be handled by the Hub
	sendJSONError(w, http.StatusNotImplemented, "WebSocket not yet implemented")
}

// handleInstances handles instance list requests
func (h *Handler) handleInstances(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		h.listInstances(w, r)
	} else {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
	}
}

// handleInstanceRouter routes /api/instances/ requests to appropriate handlers
func (h *Handler) handleInstanceRouter(w http.ResponseWriter, r *http.Request) {
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		sendJSONError(w, http.StatusBadRequest, "Invalid path")
		return
	}

	instanceKey := parts[3]

	// Check if it's a control request: /api/instances/{key}/stop or /restart
	if len(parts) >= 5 {
		action := parts[4]
		if action == "stop" {
			h.stopInstance(w, r, instanceKey)
			return
		}
		if action == "restart" {
			h.restartInstance(w, r, instanceKey)
			return
		}
	}

	// Check if it's a DELETE request for deleting offline instance
	if r.Method == http.MethodDelete {
		h.deleteOfflineInstance(w, r, instanceKey)
		return
	}

	// Otherwise, it's an instance detail request: /api/instances/{key}
	h.getInstanceDetail(w, r, instanceKey)
}

// listInstances lists all instances or instances for a specific service
func (h *Handler) listInstances(w http.ResponseWriter, r *http.Request) {
	serviceName := r.URL.Query().Get("service")

	var instances []*storage.Instance
	var err error

	if serviceName != "" {
		instances, err = h.controller.ListInstances(serviceName)
	} else {
		instances, err = h.controller.ListAllInstances()
	}

	if err != nil {
		sendJSONError(w, http.StatusInternalServerError, "Failed to get instances")
		return
	}

	sendJSON(w, instances)
}

// getInstanceDetail retrieves a specific instance
func (h *Handler) getInstanceDetail(w http.ResponseWriter, r *http.Request, instanceKey string) {
	if r.Method != http.MethodGet {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	instance, err := h.controller.GetInstance(instanceKey)
	if err != nil {
		sendJSONError(w, http.StatusNotFound, "Instance not found")
		return
	}

	sendJSON(w, instance)
}

// stopInstance stops a specific instance (POST /api/instances/{key}/stop)
func (h *Handler) stopInstance(w http.ResponseWriter, r *http.Request, instanceKey string) {
	if r.Method != http.MethodPost {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if !auth.IsAdmin(r) {
		sendJSONError(w, http.StatusForbidden, "Admin access required")
		return
	}

	if err := h.controller.StopInstance(instanceKey); err != nil {
		sendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
	sendJSON(w, map[string]string{"status": "stopping", "instance": instanceKey})
}

// restartInstance restarts a specific instance (POST /api/instances/{key}/restart)
func (h *Handler) restartInstance(w http.ResponseWriter, r *http.Request, instanceKey string) {
	if r.Method != http.MethodPost {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	if !auth.IsAdmin(r) {
		sendJSONError(w, http.StatusForbidden, "Admin access required")
		return
	}

	if err := h.controller.RestartInstance(instanceKey); err != nil {
		sendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusAccepted)
	sendJSON(w, map[string]string{"status": "restarting", "instance": instanceKey})
}

// deleteOfflineInstance deletes an offline instance (DELETE /api/instances/{key})
func (h *Handler) deleteOfflineInstance(w http.ResponseWriter, r *http.Request, instanceKey string) {
	if !auth.IsAdmin(r) {
		sendJSONError(w, http.StatusForbidden, "Admin access required")
		return
	}

	if err := h.controller.DeleteOfflineInstance(instanceKey); err != nil {
		sendJSONError(w, http.StatusInternalServerError, err.Error())
		return
	}

	w.WriteHeader(http.StatusOK)
	sendJSON(w, map[string]string{"status": "deleted", "instance": instanceKey})
}
