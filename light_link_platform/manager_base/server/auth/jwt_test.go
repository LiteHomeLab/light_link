package auth

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestGenerateAndValidateToken(t *testing.T) {
	auth := NewAuthMiddleware("test-secret", 24*time.Hour)

	token, err := auth.GenerateToken("testuser", "admin")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	if token == "" {
		t.Fatal("Token is empty")
	}

	// Validate token
	claims, err := auth.ValidateToken(token)
	if err != nil {
		t.Fatalf("ValidateToken failed: %v", err)
	}

	if claims.Username != "testuser" {
		t.Errorf("Expected username 'testuser', got '%s'", claims.Username)
	}

	if claims.Role != "admin" {
		t.Errorf("Expected role 'admin', got '%s'", claims.Role)
	}
}

func TestInvalidToken(t *testing.T) {
	auth := NewAuthMiddleware("test-secret", 24*time.Hour)

	_, err := auth.ValidateToken("invalid-token")
	if err == nil {
		t.Error("Expected error for invalid token")
	}
}

func TestMiddleware(t *testing.T) {
	auth := NewAuthMiddleware("test-secret", 24*time.Hour)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	middleware := auth.Middleware()(handler)

	// Test without token
	req := httptest.NewRequest("GET", "/api/test", nil)
	rec := httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("Expected status %d, got %d", http.StatusUnauthorized, rec.Code)
	}

	// Test with valid token
	token, _ := auth.GenerateToken("testuser", "viewer")
	req = httptest.NewRequest("GET", "/api/test", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec = httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d, got %d", http.StatusOK, rec.Code)
	}

	// Test login endpoint bypass
	req = httptest.NewRequest("POST", "/api/auth/login", nil)
	rec = httptest.NewRecorder()
	middleware.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Login endpoint should bypass auth, got status %d", rec.Code)
	}
}

func TestRequireAdmin(t *testing.T) {
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	adminHandler := RequireAdmin(handler)

	// Test without role
	req := httptest.NewRequest("GET", "/api/admin", nil)
	rec := httptest.NewRecorder()
	adminHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d, got %d", http.StatusForbidden, rec.Code)
	}

	// Test with viewer role
	req = httptest.NewRequest("GET", "/api/admin", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", "viewer"))
	rec = httptest.NewRecorder()
	adminHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusForbidden {
		t.Errorf("Expected status %d for viewer, got %d", http.StatusForbidden, rec.Code)
	}

	// Test with admin role
	req = httptest.NewRequest("GET", "/api/admin", nil)
	req = req.WithContext(context.WithValue(req.Context(), "role", "admin"))
	rec = httptest.NewRecorder()
	adminHandler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("Expected status %d for admin, got %d", http.StatusOK, rec.Code)
	}
}

func TestTokenExpiry(t *testing.T) {
	auth := NewAuthMiddleware("test-secret", 1*time.Millisecond)

	token, err := auth.GenerateToken("testuser", "viewer")
	if err != nil {
		t.Fatalf("GenerateToken failed: %v", err)
	}

	// Wait for token to expire
	time.Sleep(10 * time.Millisecond)

	_, err = auth.ValidateToken(token)
	if err == nil {
		t.Error("Expected error for expired token")
	}
}
