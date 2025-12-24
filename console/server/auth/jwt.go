package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Claims represents the JWT claims
type Claims struct {
	Username string `json:"username"`
	Role     string `json:"role"`
	jwt.RegisteredClaims
}

// AuthMiddleware handles JWT authentication
type AuthMiddleware struct {
	secret string
	expiry time.Duration
}

// NewAuthMiddleware creates a new auth middleware
func NewAuthMiddleware(secret string, expiry time.Duration) *AuthMiddleware {
	return &AuthMiddleware{
		secret: secret,
		expiry: expiry,
	}
}

// GenerateToken generates a JWT token for a user
func (a *AuthMiddleware) GenerateToken(username, role string) (string, error) {
	claims := Claims{
		Username: username,
		Role:     role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(a.expiry)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Issuer:    "light-link-console",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(a.secret))
}

// ValidateToken validates a JWT token and returns the claims
func (a *AuthMiddleware) ValidateToken(tokenString string) (*Claims, error) {
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")

	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(a.secret), nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, jwt.ErrInvalidKey
}

// Middleware returns an HTTP middleware that validates JWT tokens
func (a *AuthMiddleware) Middleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Skip auth for login endpoint
			if r.URL.Path == "/api/auth/login" {
				next.ServeHTTP(w, r)
				return
			}

			// Skip auth for WebSocket (handled separately)
			if r.URL.Path == "/api/ws" {
				next.ServeHTTP(w, r)
				return
			}

			// Get token from Authorization header
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				a.sendUnauthorized(w)
				return
			}

			// Validate token
			claims, err := a.ValidateToken(authHeader)
			if err != nil {
				a.sendUnauthorized(w)
				return
			}

			// Add user info to context
			ctx := context.WithValue(r.Context(), "username", claims.Username)
			ctx = context.WithValue(ctx, "role", claims.Role)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// RequireAdmin is middleware that requires admin role
func RequireAdmin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.Context().Value("role")
		if role != "admin" {
			sendJSONError(w, http.StatusForbidden, "Admin access required")
			return
		}
		next.ServeHTTP(w, r)
	})
}

// sendUnauthorized sends an unauthorized response
func (a *AuthMiddleware) sendUnauthorized(w http.ResponseWriter) {
	sendJSONError(w, http.StatusUnauthorized, "Unauthorized")
}

// sendJSONError sends a JSON error response
func sendJSONError(w http.ResponseWriter, status int, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]string{"error": message})
}

// GetUsername returns the username from the request context
func GetUsername(r *http.Request) string {
	if username, ok := r.Context().Value("username").(string); ok {
		return username
	}
	return ""
}

// GetRole returns the role from the request context
func GetRole(r *http.Request) string {
	if role, ok := r.Context().Value("role").(string); ok {
		return role
	}
	return ""
}

// IsAdmin checks if the current user is an admin
func IsAdmin(r *http.Request) bool {
	return GetRole(r) == "admin"
}
