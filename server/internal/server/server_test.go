package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestNewValidConfig(t *testing.T) {
	s, err := New(Config{Port: 9876}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.cfg.Port != 9876 {
		t.Errorf("expected port 9876, got %d", s.cfg.Port)
	}
}

func TestNewDefaults(t *testing.T) {
	s, err := New(Config{}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if s.cfg.Port != 9876 {
		t.Errorf("expected default port 9876, got %d", s.cfg.Port)
	}
	if s.cfg.MaxFileSize != 1048576 {
		t.Errorf("expected default max file size 1048576, got %d", s.cfg.MaxFileSize)
	}
	if s.cfg.LogLevel != "info" {
		t.Errorf("expected default log level info, got %q", s.cfg.LogLevel)
	}
}

func TestHealthHandler(t *testing.T) {
	s, err := New(Config{Port: 9876, ProjectID: "test-project"}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	time.Sleep(time.Millisecond)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	s.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response body: %v", err)
	}
	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got %v", body["status"])
	}
	if body["version"] != "0.1.0" {
		t.Errorf("expected version '0.1.0', got %v", body["version"])
	}
	if body["project_id"] != "test-project" {
		t.Errorf("expected project_id 'test-project', got %v", body["project_id"])
	}
	if body["auth_mode"] != "service_account" {
		t.Errorf("expected auth_mode 'service_account', got %v", body["auth_mode"])
	}
	uptime, ok := body["uptime_seconds"].(float64)
	if !ok || uptime < 0 {
		t.Errorf("expected positive uptime_seconds, got %v", body["uptime_seconds"])
	}
}

func TestCORSMiddleware(t *testing.T) {
	s, err := New(Config{Port: 9876}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/health", nil)
	s.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if origin := resp.Header.Get("Access-Control-Allow-Origin"); origin != "*" {
		t.Errorf("expected Access-Control-Allow-Origin: *, got %q", origin)
	}
	if methods := resp.Header.Get("Access-Control-Allow-Methods"); methods != "GET, POST, OPTIONS" {
		t.Errorf("expected Access-Control-Allow-Methods: GET, POST, OPTIONS, got %q", methods)
	}
}

func TestCORSPreflight(t *testing.T) {
	s, err := New(Config{Port: 9876}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("OPTIONS", "/health", nil)
	s.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("expected 204 No Content, got %d", resp.StatusCode)
	}
}

func TestUploadBodyLimit(t *testing.T) {
	s, err := New(Config{Port: 9876, ProjectID: "test-project", MaxFileSize: 1048576}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	body := strings.NewReader(strings.Repeat("a", 11<<20))
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", body)
	r.Header.Set("Content-Type", "application/json")
	s.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413 Payload Too Large, got %d", resp.StatusCode)
	}

	var errResp map[string]any
	json.NewDecoder(resp.Body).Decode(&errResp)
	if errResp["code"] != "PAYLOAD_TOO_LARGE" {
		t.Errorf("expected code PAYLOAD_TOO_LARGE, got %v", errResp["code"])
	}
}

func TestShutdownHandler(t *testing.T) {
	s, err := New(Config{Port: 9876}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/shutdown", nil)
	s.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200 OK, got %d", resp.StatusCode)
	}

	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if body["success"] != true {
		t.Errorf("expected success true, got %v", body["success"])
	}
	if body["status"] != "shutting_down" {
		t.Errorf("expected status 'shutting_down', got %v", body["status"])
	}
}

func TestAuthHandlerReturnsNotImplemented(t *testing.T) {
	s, err := New(Config{Port: 9876}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/auth", nil)
	s.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotImplemented {
		t.Errorf("expected 501 Not Implemented, got %d", resp.StatusCode)
	}
}

func TestShutdown(t *testing.T) {
	s, err := New(Config{Port: 9876}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	if err := s.Shutdown(ctx); err != nil {
		t.Errorf("unexpected shutdown error: %v", err)
	}
}

func Test404ForUnknownRoute(t *testing.T) {
	s, err := New(Config{Port: 9876}, nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	w := httptest.NewRecorder()
	r := httptest.NewRequest("GET", "/unknown", nil)
	s.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
}
