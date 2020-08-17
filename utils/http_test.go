package utils

import (
	"listenstats/config"
	"net/http/httptest"
	"testing"
)

var basicConfig = config.Config{
	TrustedProxies: make([]string, 0),
}

func TestRemoteAddrSimple(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost:3000", nil)
	req.RemoteAddr = "127.0.0.1:5589"
	result := FindClientRemoteAddr(&basicConfig, req)
	if result != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %s", result)
	}
}

var xffConfig = config.Config{
	TrustedProxies: []string{"127.0.0.3"},
}

func TestRemoteAddrXFF(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost:3000", nil)
	req.RemoteAddr = "127.0.0.1:5589"
	req.Header.Add("X-Forwarded-For", "127.0.0.2, 127.0.0.3")
	result := FindClientRemoteAddr(&xffConfig, req)
	if result != "127.0.0.2" {
		t.Fatalf("expected 127.0.0.2, got %s", result)
	}
}

func TestRemoteAddrXFFUntrusted(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost:3000", nil)
	req.RemoteAddr = "127.0.0.1:5589"
	req.Header.Add("X-Forwarded-For", "127.0.0.2, 127.0.0.4")
	result := FindClientRemoteAddr(&xffConfig, req)
	if result != "127.0.0.1" {
		t.Fatalf("expected 127.0.0.1, got %s", result)
	}
}

func TestRemoteAddrXFFMulti(t *testing.T) {
	req := httptest.NewRequest("GET", "http://localhost:3000", nil)
	req.RemoteAddr = "127.0.0.1:5589"
	req.Header.Add("X-Forwarded-For", "127.0.0.2, 127.0.0.4, 127.0.0.3")
	result := FindClientRemoteAddr(&xffConfig, req)
	if result != "127.0.0.2" {
		t.Fatalf("expected 127.0.0.2, got %s", result)
	}
}
