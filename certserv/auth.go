package certserv

import (
	"fmt"
	"net/http"
	"os"
)

// BasicAuthMiddleware provides HTTP Basic Authentication middleware
func BasicAuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get credentials from environment or use defaults
		expectedUser := os.Getenv("ADCS_AUTH_USER")
		expectedPass := os.Getenv("ADCS_AUTH_PASSWORD")

		// If not set in environment, use defaults (should be changed in production)
		if expectedUser == "" {
			expectedUser = "admin"
		}
		if expectedPass == "" {
			expectedPass = "changeme"
		}

		// Check if Basic Auth header is present
		user, pass, ok := r.BasicAuth()
		if !ok {
			w.Header().Set("WWW-Authenticate", `Basic realm="ADCS Simulator"`)
			http.Error(w, "Unauthorized: Basic auth required", http.StatusUnauthorized)
			return
		}

		// Verify credentials
		if user != expectedUser || pass != expectedPass {
			w.Header().Set("WWW-Authenticate", `Basic realm="ADCS Simulator"`)
			http.Error(w, "Unauthorized: Invalid credentials", http.StatusUnauthorized)
			setupLog.Warn("Failed authentication attempt", "user", user)
			return
		}

		setupLog.Debug("Authenticated request", "user", user)
		next(w, r)
	}
}

// AuthStatusHandler returns the current authentication status
func AuthStatusHandler(w http.ResponseWriter, r *http.Request) {
	user, _, ok := r.BasicAuth()
	if !ok {
		w.Header().Set("WWW-Authenticate", `Basic realm="ADCS Simulator"`)
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	fmt.Fprintf(w, `{"authenticated":true,"user":"%s"}`, user)
}
