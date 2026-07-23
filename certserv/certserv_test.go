package certserv

import (
	"strings"
	"testing"
	"time"
	"net/http"
	"net/http/httptest"
)

func TestGetEnv(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		expected string
		setup    func()
		cleanup  func()
	}{
		{
			name:     "key exists",
			key:      "TEST_KEY_EXISTS",
			fallback: "fallback",
			expected: "test_value",
			setup: func() {
				t.Setenv("TEST_KEY_EXISTS", "test_value")
			},
		},
		{
			name:     "key not exists",
			key:      "TEST_KEY_NOT_EXISTS_XYZ",
			fallback: "fallback_value",
			expected: "fallback_value",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setup != nil {
				tt.setup()
			}
			result := getEnv(tt.key, tt.fallback)
			if result != tt.expected {
				t.Errorf("getEnv(%s, %s) = %s, want %s", tt.key, tt.fallback, result, tt.expected)
			}
		})
	}
}

func TestCertTimeToSign(t *testing.T) {
	tests := []struct {
		name     string
		signTime time.Time
		expected bool
	}{
		{
			name:     "time in past",
			signTime: time.Now().Add(-1 * time.Hour),
			expected: true,
		},
		{
			name:     "time in future",
			signTime: time.Now().Add(1 * time.Hour),
			expected: false,
		},
		{
			name:     "time is now",
			signTime: time.Now(),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cert := &Cert{SignTime: tt.signTime}
			result := cert.TimeToSign()
			if result != tt.expected {
				t.Errorf("TimeToSign() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestRespondError(t *testing.T) {
	tests := []struct {
		name       string
		text       string
		wantStatus int
		wantBody   string
	}{
		{
			name:       "error message",
			text:       "test error",
			wantStatus: http.StatusBadRequest,
			wantBody:   "test error\n",
		},
		{
			name:       "empty error",
			text:       "",
			wantStatus: http.StatusBadRequest,
			wantBody:   "\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondError(w, tt.text)

			if w.Code != tt.wantStatus {
				t.Errorf("respondError() status = %d, want %d", w.Code, tt.wantStatus)
			}
			if w.Body.String() != tt.wantBody {
				t.Errorf("respondError() body = %q, want %q", w.Body.String(), tt.wantBody)
			}
		})
	}
}

func TestGetSimOrders(t *testing.T) {
	tests := []struct {
		name            string
		dnsNames        []string
		expectReject    bool
		expectUnauth    bool
		expectDelayMin  time.Duration
	}{
		{
			name:            "no special names",
			dnsNames:        []string{"example.com", "test.com"},
			expectReject:    false,
			expectUnauth:    false,
			expectDelayMin:  0,
		},
		{
			name:            "reject domain",
			dnsNames:        []string{"reject.sim", "example.com"},
			expectReject:    true,
			expectUnauth:    false,
			expectDelayMin:  0,
		},
		{
			name:            "unauthorized domain",
			dnsNames:        []string{"unauthorized.sim", "example.com"},
			expectReject:    false,
			expectUnauth:    true,
			expectDelayMin:  0,
		},
		{
			name:            "delay domain",
			dnsNames:        []string{"delay.5s.sim", "example.com"},
			expectReject:    false,
			expectUnauth:    false,
			expectDelayMin:  5 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders := getSimOrders(tt.dnsNames)

			if orders.reject != tt.expectReject {
				t.Errorf("getSimOrders() reject = %v, want %v", orders.reject, tt.expectReject)
			}
			if orders.unauthorized != tt.expectUnauth {
				t.Errorf("getSimOrders() unauthorized = %v, want %v", orders.unauthorized, tt.expectUnauth)
			}
			if orders.delay != tt.expectDelayMin {
				t.Errorf("getSimOrders() delay = %v, want %v", orders.delay, tt.expectDelayMin)
			}
		})
	}
}

func TestDecodeCertRequest(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "invalid pem",
			data:      "not a pem block",
			wantError: true,
			errorMsg:  "Cannot decode CSR PEM",
		},
		{
			name:      "empty data",
			data:      "",
			wantError: true,
			errorMsg:  "Cannot decode CSR PEM",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeCertRequest(tt.data)

			if (err != nil) != tt.wantError {
				t.Errorf("decodeCertRequest() error = %v, wantError %v", err, tt.wantError)
			}
			if err != nil && tt.errorMsg != "" && !strings.Contains(err.Error(), tt.errorMsg) {
				t.Errorf("decodeCertRequest() error message = %q, want to contain %q", err.Error(), tt.errorMsg)
			}
		})
	}
}

func TestBasicAuthMiddleware(t *testing.T) {
	tests := []struct {
		name           string
		setCredentials bool
		username       string
		password       string
		expectedStatus int
		setupEnv       func()
		cleanupEnv     func()
	}{
		{
			name:           "no credentials",
			setCredentials: false,
			expectedStatus: http.StatusUnauthorized,
			setupEnv: func() {
				t.Setenv("ADCS_AUTH_USER", "")
				t.Setenv("ADCS_AUTH_PASSWORD", "")
			},
		},
		{
			name:           "invalid credentials",
			setCredentials: true,
			username:       "wronguser",
			password:       "wrongpass",
			expectedStatus: http.StatusUnauthorized,
			setupEnv: func() {
				t.Setenv("ADCS_AUTH_USER", "admin")
				t.Setenv("ADCS_AUTH_PASSWORD", "changeme")
			},
		},
		{
			name:           "valid credentials",
			setCredentials: true,
			username:       "admin",
			password:       "changeme",
			expectedStatus: http.StatusOK,
			setupEnv: func() {
				t.Setenv("ADCS_AUTH_USER", "admin")
				t.Setenv("ADCS_AUTH_PASSWORD", "changeme")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.setupEnv != nil {
				tt.setupEnv()
			}

			nextCalled := false
			handler := BasicAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			if tt.setCredentials {
				req.SetBasicAuth(tt.username, tt.password)
			}
			w := httptest.NewRecorder()

			handler(w, req)

			if w.Code != tt.expectedStatus {
				t.Errorf("BasicAuthMiddleware() status = %d, want %d", w.Code, tt.expectedStatus)
			}

			if tt.expectedStatus == http.StatusOK && !nextCalled {
				t.Error("BasicAuthMiddleware() next handler not called for valid credentials")
			}
		})
	}
}
