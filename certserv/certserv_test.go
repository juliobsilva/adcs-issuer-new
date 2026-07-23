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
		{"past time", time.Now().Add(-1 * time.Hour), true},
		{"future time", time.Now().Add(1 * time.Hour), false},
		{"current time", time.Now(), true},
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
	}{
		{"error message", "test error", http.StatusBadRequest},
		{"empty error", "", http.StatusBadRequest},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			respondError(w, tt.text)
			if w.Code != tt.wantStatus {
				t.Errorf("respondError() status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}

func TestGetSimOrders(t *testing.T) {
	tests := []struct {
		name         string
		dnsNames     []string
		expectReject bool
	}{
		{"no special", []string{"example.com"}, false},
		{"reject", []string{"reject.sim"}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			orders := getSimOrders(tt.dnsNames)
			if orders == nil {
				t.Fatal("getSimOrders returned nil")
			}
			if orders.reject != tt.expectReject {
				t.Errorf("reject = %v, want %v", orders.reject, tt.expectReject)
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
		{"invalid pem", "not pem", true},
		{"empty data", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := decodeCertRequest(tt.data)
			if (err != nil) != tt.wantError {
				t.Errorf("error = %v, wantError %v", err, tt.wantError)
			}
		})
	}
}

func TestBasicAuthMiddleware(t *testing.T) {
	t.Setenv("ADCS_AUTH_USER", "user")
	t.Setenv("ADCS_AUTH_PASSWORD", "pass")

	tests := []struct {
		name   string
		user   string
		pass   string
		status int
	}{
		{"valid", "user", "pass", http.StatusOK},
		{"invalid", "wrong", "wrong", http.StatusUnauthorized},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			called := false
			h := BasicAuthMiddleware(func(w http.ResponseWriter, r *http.Request) {
				called = true
				w.WriteHeader(http.StatusOK)
			})

			req := httptest.NewRequest("GET", "/", nil)
			req.SetBasicAuth(tt.user, tt.pass)
			w := httptest.NewRecorder()

			h(w, req)

			if w.Code != tt.status {
				t.Errorf("status = %d, want %d", w.Code, tt.status)
			}
			if tt.status == http.StatusOK && !called {
				t.Error("handler not called")
			}
		})
	}
}

func TestCertStruct(t *testing.T) {
	now := time.Now()
	cert := &Cert{Crt: "cert", Csr: "csr", SignTime: now}
	if cert.Crt != "cert" {
		t.Errorf("Cert.Crt = %q, want cert", cert.Crt)
	}
}

func TestSimOrdersStruct(t *testing.T) {
	orders := &SimOrders{reject: true}
	if !orders.reject {
		t.Error("SimOrders.reject should be true")
	}
}


