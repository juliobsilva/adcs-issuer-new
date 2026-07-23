package certserv

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
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
	// Test basic functionality of getSimOrders
	tests := []struct {
		name         string
		dnsNames     []string
		expectReject bool
	}{
		{
			name:         "no special names",
			dnsNames:     []string{"example.com", "test.com"},
			expectReject: false,
		},
		{
			name:         "reject domain",
			dnsNames:     []string{"reject.sim", "example.com"},
			expectReject: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders := getSimOrders(tt.dnsNames)
			if orders == nil {
				t.Fatal("getSimOrders() returned nil")
			}
			if orders.reject != tt.expectReject {
				t.Errorf("getSimOrders() reject = %v, want %v", orders.reject, tt.expectReject)
			}
		})
	}
}

func TestDecodeCertRequest(t *testing.T) {
	tests := []struct {
		name      string
		data      string
		wantError bool
	}{
		{
			name:      "invalid pem",
			data:      "not a pem block",
			wantError: true,
		},
		{
			name:      "empty data",
			data:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeCertRequest(tt.data)
			if (err != nil) != tt.wantError {
				t.Errorf("decodeCertRequest() error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestBasicAuthMiddleware(t *testing.T) {
	// Set up environment for testing
	t.Setenv("ADCS_AUTH_USER", "testuser")
	t.Setenv("ADCS_AUTH_PASSWORD", "testpass")

	tests := []struct {
		name           string
		username       string
		password       string
		expectedStatus int
	}{
		{
			name:           "valid credentials",
			username:       "testuser",
			password:       "testpass",
			expectedStatus: http.StatusOK,
		},
		{
			name:           "invalid credentials",
			username:       "wronguser",
			password:       "wrongpass",
			expectedStatus: http.StatusUnauthorized,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			nextCalled := false
			handler := BasicAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				nextCalled = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/test", nil)
			req.SetBasicAuth(tt.username, tt.password)
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

func TestCertStruct(t *testing.T) {
	// Test Cert struct initialization and methods
	now := time.Now()
	cert := &Cert{
		Crt:      "test_cert",
		Csr:      "test_csr",
		Deny:     false,
		Denied:   false,
		SignTime: now,
	}

	if cert.Crt != "test_cert" {
		t.Errorf("Cert.Crt = %q, want %q", cert.Crt, "test_cert")
	}

	if cert.TimeToSign() == false {
		t.Error("TimeToSign() should return true for current time")
	}
}

func TestSimOrdersStruct(t *testing.T) {
	// Test SimOrders struct
	orders := &SimOrders{
		reject:       true,
		delay:        100 * time.Millisecond,
		unauthorized: false,
	}

	if !orders.reject {
		t.Error("SimOrders.reject should be true")
	}

	if orders.delay != 100*time.Millisecond {
		t.Errorf("SimOrders.delay = %v, want 100ms", orders.delay)
	}

	if orders.unauthorized {
		t.Error("SimOrders.unauthorized should be false")
	}
}
