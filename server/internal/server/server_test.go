package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestNew(t *testing.T) {
	s := New(8080)
	if s.port != 8080 {
		t.Errorf("expected port 8080, got %d", s.port)
	}
}

func TestHealthHandler(t *testing.T) {
	s := New(0)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)

	s.HealthHandler(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", body["status"])
	}
	if body["version"] != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %v", body["version"])
	}
}
