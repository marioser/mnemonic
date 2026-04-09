package http

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/marioser/mnemonic/internal/config"
)

func TestHealthEndpoint(t *testing.T) {
	cfg := config.Default()
	s := New(cfg, nil) // svc nil is ok for health check

	req := httptest.NewRequest("GET", "/health", nil)
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["service"] != "mnemonic" {
		t.Errorf("expected service=mnemonic, got %s", resp["service"])
	}
	if resp["status"] != "ok" {
		t.Errorf("expected status=ok, got %s", resp["status"])
	}
}

func TestSessionEndEndpoint(t *testing.T) {
	cfg := config.Default()
	s := New(cfg, nil)

	req := httptest.NewRequest("POST", "/session/end", nil)
	w := httptest.NewRecorder()

	s.Router().ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", w.Code)
	}

	var resp map[string]string
	json.NewDecoder(w.Body).Decode(&resp)

	if resp["status"] != "acknowledged" {
		t.Errorf("expected acknowledged, got %s", resp["status"])
	}
}
