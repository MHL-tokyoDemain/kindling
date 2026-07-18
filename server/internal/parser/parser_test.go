package parser

import (
	"testing"

	"github.com/kindling/kindling/pkg/types"
)

func TestParseJSONObject(t *testing.T) {
	content := []byte(`{"name": "test", "value": 42}`)
	result := Parse("data.json", content, 1048576)

	if result.Code != "" {
		t.Errorf("expected no error code, got %q", result.Code)
	}
	if result.Error != "" {
		t.Errorf("expected no error, got %q", result.Error)
	}
	if result.Document == nil {
		t.Fatal("expected document, got nil")
	}
	if result.Document.ContentType != "json" {
		t.Errorf("expected content_type json, got %q", result.Document.ContentType)
	}

	obj, ok := result.Document.Content.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any content, got %T", result.Document.Content)
	}
	if obj["name"] != "test" {
		t.Errorf("expected name=test, got %v", obj["name"])
	}
	if obj["value"] != float64(42) {
		t.Errorf("expected value=42, got %v", obj["value"])
	}
}

func TestParseJSONArray(t *testing.T) {
	content := []byte(`[1, 2, 3]`)
	result := Parse("data.json", content, 1048576)

	if result.Code != types.ErrParseFailed {
		t.Errorf("expected code PARSE_001, got %q", result.Code)
	}
	if result.Error != "expected JSON object, got array" {
		t.Errorf("unexpected error: %q", result.Error)
	}
	if result.Document != nil {
		t.Error("expected nil document for array")
	}
}

func TestParseJSONPrimitive(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
	}{
		{"number", []byte(`42`)},
		{"string", []byte(`"hello"`)},
		{"bool", []byte(`true`)},
		{"null", []byte(`null`)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Parse("data.json", tt.content, 1048576)

			if result.Code != types.ErrParseFailed {
				t.Errorf("expected code PARSE_001, got %q", result.Code)
			}
			if result.Error != "expected JSON object" {
				t.Errorf("unexpected error: %q", result.Error)
			}
			if result.Document != nil {
				t.Error("expected nil document")
			}
		})
	}
}

func TestParseJSONSyntaxError(t *testing.T) {
	content := []byte(`{invalid}`)
	result := Parse("data.json", content, 1048576)

	if result.Code != types.ErrParseFailed {
		t.Errorf("expected code PARSE_001, got %q", result.Code)
	}
	if result.Error == "" || result.Error[:14] != "invalid JSON: " {
		t.Errorf("expected error starting with 'invalid JSON: ', got %q", result.Error)
	}
	if result.Document != nil {
		t.Error("expected nil document for invalid JSON")
	}
}

func TestParseEmptyJSON(t *testing.T) {
	content := []byte{}
	result := Parse("data.json", content, 1048576)

	if result.Code != types.ErrEmptyJSON {
		t.Errorf("expected code EMPTY_JSON, got %q", result.Code)
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
	if result.Document != nil {
		t.Error("expected nil document for empty JSON")
	}
}

func TestParseTextFile(t *testing.T) {
	content := []byte("Hello, world!")
	result := Parse("log.txt", content, 1048576)

	if result.Code != "" {
		t.Errorf("expected no error code, got %q", result.Code)
	}
	if result.Error != "" {
		t.Errorf("expected no error, got %q", result.Error)
	}
	if result.Document == nil {
		t.Fatal("expected document, got nil")
	}
	if result.Document.ContentType != "text" {
		t.Errorf("expected content_type text, got %q", result.Document.ContentType)
	}

	str, ok := result.Document.Content.(string)
	if !ok {
		t.Fatalf("expected string content, got %T", result.Document.Content)
	}
	if str != "Hello, world!" {
		t.Errorf("expected 'Hello, world!', got %q", str)
	}
}

func TestParseEmptyText(t *testing.T) {
	content := []byte{}
	result := Parse("empty.txt", content, 1048576)

	if result.Code != "" {
		t.Errorf("expected no error code, got %q", result.Code)
	}
	if result.Error != "" {
		t.Errorf("expected no error, got %q", result.Error)
	}
	if result.Document == nil {
		t.Fatal("expected document, got nil")
	}

	str, ok := result.Document.Content.(string)
	if !ok {
		t.Fatalf("expected string content, got %T", result.Document.Content)
	}
	if str != "" {
		t.Errorf("expected empty string, got %q", str)
	}
}

func TestParseTextWithBOM(t *testing.T) {
	content := []byte{0xEF, 0xBB, 0xBF, 'h', 'e', 'l', 'l', 'o'}
	result := Parse("log.txt", content, 1048576)

	str, ok := result.Document.Content.(string)
	if !ok {
		t.Fatalf("expected string content, got %T", result.Document.Content)
	}
	if str != "hello" {
		t.Errorf("expected 'hello' (BOM stripped), got %q", str)
	}
}

func TestParseExceedsMaxFileSize(t *testing.T) {
	content := []byte("hello")
	result := Parse("file.txt", content, 2)

	if result.Code != types.ErrPayloadTooLarge {
		t.Errorf("expected code PAYLOAD_TOO_LARGE, got %q", result.Code)
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
	if result.Document != nil {
		t.Error("expected nil document for oversized file")
	}
}

func TestParseExactMaxFileSize(t *testing.T) {
	content := []byte("hello")
	result := Parse("file.txt", content, 5)

	if result.Code != "" {
		t.Errorf("expected no error, got code %q", result.Code)
	}
	if result.Document == nil {
		t.Fatal("expected document for file at max size")
	}
}

func TestParseUnknownExtension(t *testing.T) {
	content := []byte("some data")
	result := Parse("data.csv", content, 1048576)

	if result.Code != types.ErrInvalidContentType {
		t.Errorf("expected code INVALID_CONTENT_TYPE, got %q", result.Code)
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
	if result.Document != nil {
		t.Error("expected nil document for unknown extension")
	}
}

func TestParseMixedJSON(t *testing.T) {
	content := []byte(`{"nested": {"a": 1, "b": [2, 3]}}`)
	result := Parse("data.json", content, 1048576)

	if result.Code != "" {
		t.Errorf("expected no error code, got %q", result.Code)
	}
	if result.Document == nil {
		t.Fatal("expected document, got nil")
	}

	obj, ok := result.Document.Content.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result.Document.Content)
	}
	nested, ok := obj["nested"].(map[string]any)
	if !ok {
		t.Fatalf("expected nested map, got %T", obj["nested"])
	}
	if nested["a"] != float64(1) {
		t.Errorf("expected a=1, got %v", nested["a"])
	}
}

func TestParseNoExtension(t *testing.T) {
	content := []byte("some data")
	result := Parse("Makefile", content, 1048576)

	if result.Code != types.ErrInvalidContentType {
		t.Errorf("expected code INVALID_CONTENT_TYPE, got %q", result.Code)
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
	if result.Document != nil {
		t.Error("expected nil document for unknown extension")
	}
}

func TestParseUpperCaseExtension(t *testing.T) {
	content := []byte(`{"key": "value"}`)
	result := Parse("DATA.JSON", content, 1048576)

	if result.Code != "" {
		t.Errorf("expected no error for .JSON extension, got code %q", result.Code)
	}
	if result.Document == nil {
		t.Fatal("expected document, got nil")
	}
	if result.Document.ContentType != "json" {
		t.Errorf("expected content_type json, got %q", result.Document.ContentType)
	}
}

func TestParseMultiLineText(t *testing.T) {
	content := []byte("line 1\nline 2\nline 3")
	result := Parse("notes.txt", content, 1048576)

	str, ok := result.Document.Content.(string)
	if !ok {
		t.Fatalf("expected string content, got %T", result.Document.Content)
	}
	if str != "line 1\nline 2\nline 3" {
		t.Errorf("unexpected content: %q", str)
	}
}

func TestParseJSONWithBOM(t *testing.T) {
	content := []byte{0xEF, 0xBB, 0xBF, '{', '"', 'a', '"', ':', '1', '}'}
	result := Parse("data.json", content, 1048576)

	if result.Code != "" {
		t.Errorf("expected no error code, got %q", result.Code)
	}
	if result.Document == nil {
		t.Fatal("expected document, got nil")
	}
	obj, ok := result.Document.Content.(map[string]any)
	if !ok {
		t.Fatalf("expected map[string]any, got %T", result.Document.Content)
	}
	if obj["a"] != float64(1) {
		t.Errorf("expected a=1, got %v", obj["a"])
	}
}

func TestParseJSONOnlyBOM(t *testing.T) {
	content := []byte{0xEF, 0xBB, 0xBF}
	result := Parse("data.json", content, 1048576)

	if result.Code != types.ErrEmptyJSON {
		t.Errorf("expected code EMPTY_JSON, got %q", result.Code)
	}
}

func TestParseOnlyBOM(t *testing.T) {
	content := []byte{0xEF, 0xBB, 0xBF}
	result := Parse("empty.txt", content, 1048576)

	if result.Code != "" {
		t.Errorf("expected no error code, got %q", result.Code)
	}
	str, ok := result.Document.Content.(string)
	if !ok {
		t.Fatalf("expected string content, got %T", result.Document.Content)
	}
	if str != "" {
		t.Errorf("expected empty string after stripping BOM, got %q", str)
	}
}
