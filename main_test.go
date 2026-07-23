package main

import (
	"net"
	"testing"
)

func TestPtrString(t *testing.T) {
	s := "test"
	p := ptrString(s)
	if p == nil {
		t.Fatal("ptrString returned nil")
	}
	if *p != "test" {
		t.Errorf("*ptrString(\"test\") = %q, want \"test\"", *p)
	}
}

func TestNetParseIP(t *testing.T) {
	tests := []struct {
		name    string
		ip      string
		isValid bool
	}{
		{"ipv4", "127.0.0.1", true},
		{"invalid", "not-ip", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := net.ParseIP(tt.ip)
			if (result != nil) != tt.isValid {
				t.Errorf("ParseIP(%q) valid=%v, want %v", tt.ip, result != nil, tt.isValid)
			}
		})
	}
}

// Helper function
func ptrString(s string) *string {
	return &s
}
