package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/LiteHomeLab/light_link/console/server/auth"
	"github.com/LiteHomeLab/light_link/console/server/manager"
	"github.com/LiteHomeLab/light_link/console/server/storage"
)

// Handler handles API requests
type Handler struct {
	db      *storage.Database
	manager *manager.Manager
	auth    *auth.AuthMiddleware
}

// NewHandler creates a new API handler
func NewHandler(db *storage.Database, mgr *manager.Manager, auth *auth.AuthMiddleware) *Handler {
	return &Handler{
		db:      db,
		manager: mgr,
		auth:    auth,
	}
}

// Routes returns the HTTP handler with all routes registered
func (h *Handler) Routes() http.Handler {
	mux := http.NewServeMux()

	// Auth endpoints
	mux.HandleFunc("/api/auth/login", h.handleLogin)

	// Service endpoints
	mux.HandleFunc("/api/services", h.withAuth(h.handleServices))
	mux.HandleFunc("/api/services/", h.withAuth(h.handleServiceDetail))

	// Status endpoints
	mux.HandleFunc("/api/status", h.withAuth(h.handleStatus))
	mux.HandleFunc("/api/status/", h.withAuth(h.handleServiceStatus))

	// Method endpoints
	mux.HandleFunc("/api/services/", h.withAuth(h.handleMethods))

	// Event endpoints
	mux.HandleFunc("/api/events", h.withAuth(h.handleEvents))

	// Call endpoint
	mux.HandleFunc("/api/call", h.withAuth(h.handleCall))

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

// handleServiceDetail handles service detail requests
func (h *Handler) handleServiceDetail(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		sendJSONError(w, http.StatusMethodNotAllowed, "Method not allowed")
		return
	}

	// Extract service name from path
	// /api/services/demo-service -> demo-service
	parts := strings.Split(r.URL.Path, "/")
	if len(parts) < 4 {
		sendJSONError(w, http.StatusBadRequest, "Invalid service name")
		return
	}
	serviceName := parts[3]

	service, err := h.db.GetService(serviceName)
	if err != nil {
		sendJSONError(w, http.StatusNotFound, "Service not found")
		return
	}

	sendJSON(w, service)
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

	// TODO: Implement actual RPC call via NATS
	// For now, return a mock response
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": false,
		"error":   "RPC call not yet implemented - use NATS client directly",
	})
}

// handleWebSocket handles WebSocket connections
func (h *Handler) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement WebSocket upgrade
	// This will be handled by the Hub
	sendJSONError(w, http.StatusNotImplemented, "WebSocket not yet implemented")
}
