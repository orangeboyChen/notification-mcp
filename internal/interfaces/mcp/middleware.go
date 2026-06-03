package mcp

import (
	"net/http"
	"strings"
)

// AuthMiddleware returns an HTTP middleware that validates Bearer token authentication.
// If token is empty, it returns the handler as-is (no auth).
func AuthMiddleware(token string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		if token == "" {
			return next
		}

		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"missing Authorization header"}`, http.StatusUnauthorized)
				return
			}

			// Expect "Bearer <token>" format
			const prefix = "Bearer "
			if !strings.HasPrefix(authHeader, prefix) {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"invalid Authorization format, expected Bearer token"}`, http.StatusUnauthorized)
				return
			}

			providedToken := strings.TrimPrefix(authHeader, prefix)
			if providedToken != token {
				w.Header().Set("Content-Type", "application/json")
				http.Error(w, `{"error":"invalid token"}`, http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
