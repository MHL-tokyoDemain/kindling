package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kindling/kindling/pkg/types"
)

func TestWriteJSON(t *testing.T) {
	w := httptest.NewRecorder()
	WriteJSON(w, http.StatusCreated, map[string]string{"key": "value"})

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("expected 201, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body["key"] != "value" {
		t.Errorf("expected value, got %v", body["key"])
	}
}

func TestWriteError(t *testing.T) {
	w := httptest.NewRecorder()
	WriteError(w, types.ErrCollectionRequired, "collection path is required")

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); ct != "application/json" {
		t.Errorf("expected Content-Type application/json, got %q", ct)
	}

	var body types.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode: %v", err)
	}
	if body.Success != false {
		t.Errorf("expected success false, got %v", body.Success)
	}
	if body.Code != types.ErrCollectionRequired {
		t.Errorf("expected code %q, got %q", types.ErrCollectionRequired, body.Code)
	}
	if body.Error != "collection path is required" {
		t.Errorf("unexpected error message: %q", body.Error)
	}
}

func TestErrorHTTPStatusMapping(t *testing.T) {
	tests := []struct {
		code     string
		expected int
	}{
		{types.ErrCollectionRequired, http.StatusBadRequest},
		{types.ErrCollectionProjectMismatch, http.StatusBadRequest},
		{types.ErrInvalidContentType, http.StatusBadRequest},
		{types.ErrAuthInvalidToken, http.StatusUnauthorized},
		{types.ErrAuthSessionInvalid, http.StatusUnauthorized},
		{types.ErrPayloadTooLarge, http.StatusRequestEntityTooLarge},
		{types.ErrFirestoreQuota, http.StatusTooManyRequests},
		{types.ErrParseFailed, http.StatusBadRequest},
		{types.ErrEmptyJSON, http.StatusBadRequest},
		{types.ErrInternal, http.StatusInternalServerError},
		{"UNKNOWN_CODE", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			got := ErrorHTTPStatus(tt.code)
			if got != tt.expected {
				t.Errorf("ErrorHTTPStatus(%q) = %d, want %d", tt.code, got, tt.expected)
			}
		})
	}
}

func TestWriteErrorRoundTrip(t *testing.T) {
	tests := []struct {
		code     string
		message  string
		expected int
	}{
		{types.ErrCollectionRequired, "collection required", http.StatusBadRequest},
		{types.ErrAuthInvalidToken, "bad token", http.StatusUnauthorized},
		{types.ErrPayloadTooLarge, "too big", http.StatusRequestEntityTooLarge},
		{types.ErrFirestoreQuota, "quota", http.StatusTooManyRequests},
		{types.ErrParseFailed, "parse error", http.StatusBadRequest},
		{types.ErrInternal, "internal", http.StatusInternalServerError},
	}

	for _, tt := range tests {
		t.Run(tt.code, func(t *testing.T) {
			w := httptest.NewRecorder()
			WriteError(w, tt.code, tt.message)

			resp := w.Result()
			defer resp.Body.Close()

			if resp.StatusCode != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, resp.StatusCode)
			}

			var body types.ErrorResponse
			if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
				t.Fatalf("failed to decode: %v", err)
			}
			if body.Success != false {
				t.Errorf("expected success false, got %v", body.Success)
			}
			if body.Code != tt.code {
				t.Errorf("expected code %q, got %q", tt.code, body.Code)
			}
			if body.Error != tt.message {
				t.Errorf("expected error %q, got %q", tt.message, body.Error)
			}
		})
	}
}

func TestWriteErrorInternalGenericMessage(t *testing.T) {
	w := httptest.NewRecorder()
	detail := "connection refused: dial tcp 127.0.0.1:8080: connect: connection refused"
	WriteError(w, types.ErrInternal, detail)

	resp := w.Result()
	defer resp.Body.Close()

	var body types.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}

	if body.Error != detail {
		t.Errorf("expected error message to match input, got %q", body.Error)
	}
}
