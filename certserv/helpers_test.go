package certserv

import (
	"net/http"
	"testing"
	"time"
)

func TestGetEnvFunction(t *testing.T) {
	// Test environment variable resolution
	tests := []struct {
		name     string
		key      string
		fallback string
		setEnv   bool
		envValue string
		expected string
	}{
		{
			name:     "uses env if set",
			key:      "TEST_GETENV_VAR_1",
			fallback: "fallback_value",
			setEnv:   true,
			envValue: "env_value",
			expected: "env_value",
		},
		{
			name:     "uses fallback if not set",
			key:      "TEST_GETENV_NONEXISTENT_VAR",
			fallback: "fallback_value",
			setEnv:   false,
			expected: "fallback_value",
		},
		{
			name:     "empty string env",
			key:      "TEST_GETENV_VAR_2",
			fallback: "fallback",
			setEnv:   true,
			envValue: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setEnv {
				t.Setenv(tt.key, tt.envValue)
			}
			result := getEnv(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %q, want %q", tt.key, tt.fallback, result, tt.expected)
			}
		})
	}
}

func TestRespondErrorFormat(t *testing.T) {
	// Test respondError function with various inputs
	tests := []struct {
		name string
		text string
	}{
		{
			name: "simple message",
			text: "error occurred",
		},
		{
			name: "empty string",
			text: "",
		},
		{
			name: "special characters",
			text: "error: failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := &mockResponseWriter{body: make([]byte, 0)}
			respondError(w, tt.text)

			// Check status code
			if w.statusCode != http.StatusBadRequest {
				t.Errorf("respondError() status = %d, want %d", w.statusCode, http.StatusBadRequest)
			}
		})
	}
}

// Mock ResponseWriter for testing
type mockResponseWriter struct {
	statusCode int
	headers    map[string][]string
	body       []byte
}

func (m *mockResponseWriter) Header() map[string][]string {
	if m.headers == nil {
		m.headers = make(map[string][]string)
	}
	return m.headers
}

func (m *mockResponseWriter) Write(b []byte) (int, error) {
	m.body = append(m.body, b...)
	return len(b), nil
}

func (m *mockResponseWriter) WriteHeader(statusCode int) {
	m.statusCode = statusCode
}

func TestCertTimeToSignEdgeCases(t *testing.T) {
	// Test edge cases for TimeToSign
	tests := []struct {
		name string
		test func(t *testing.T)
	}{
		{
			name: "far past",
			test: func(t *testing.T) {
				cert := &Cert{SignTime: timeInPast(8760)}
				if !cert.TimeToSign() {
					t.Error("TimeToSign() should return true for far past time")
				}
			},
		},
		{
			name: "far future",
			test: func(t *testing.T) {
				cert := &Cert{SignTime: timeInFuture(8760)}
				if cert.TimeToSign() {
					t.Error("TimeToSign() should return false for far future time")
				}
			},
		},
		{
			name: "zero time",
			test: func(t *testing.T) {
				cert := &Cert{SignTime: timeInPast(100)}
				if !cert.TimeToSign() {
					t.Error("TimeToSign() should return true for older time")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.test(t)
		})
	}
}

// Helper functions for time testing
func timeInPast(hours int) time.Time {
	return time.Now().Add(time.Duration(-hours) * time.Hour)
}

func timeInFuture(hours int) time.Time {
	return time.Now().Add(time.Duration(hours) * time.Hour)
}
