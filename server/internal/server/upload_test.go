package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	firestorepkg "github.com/kindling/kindling/internal/firestore"
	"github.com/kindling/kindling/internal/parser"
	"github.com/kindling/kindling/pkg/types"
)

// --- mock firestoreWriter ---

type mockWriter struct {
	writeFunc func(ctx context.Context, collection string, data any) (string, error)
	closeFunc func() error
	projectID string
}

func (m *mockWriter) WriteDocument(ctx context.Context, collection string, data any) (string, error) {
	return m.writeFunc(ctx, collection, data)
}

func (m *mockWriter) Close() error {
	if m.closeFunc != nil {
		return m.closeFunc()
	}
	return nil
}

func (m *mockWriter) ProjectID() string { return m.projectID }

func successWriter() *mockWriter {
	n := 0
	return &mockWriter{
		writeFunc: func(_ context.Context, _ string, _ any) (string, error) {
			n++
			return fmt.Sprintf("doc-%d", n), nil
		},
	}
}

func errorWriter(err error) *mockWriter {
	return &mockWriter{
		writeFunc: func(_ context.Context, _ string, _ any) (string, error) {
			return "", err
		},
	}
}

// --- test helpers ---

const boundProject = "my-project"
const validCollection = "projects/my-project/databases/(default)/documents/sensors"

func uploadCfg() Config {
	return Config{
		Port:        9876,
		ProjectID:   boundProject,
		MaxFileSize: 10 << 20, // 10 MB
		Concurrency: 2,
	}
}

func doUpload(t *testing.T, fw firestoreWriter, requestBody string) *http.Response {
	t.Helper()
	h := HandleUpload(fw, parser.Parse, uploadCfg(), nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(requestBody))
	r.Header.Set("Content-Type", "application/json")
	h.ServeHTTP(w, r)
	return w.Result()
}

func decodeUploadResponse(t *testing.T, resp *http.Response) types.UploadResponse {
	t.Helper()
	var out types.UploadResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode UploadResponse: %v", err)
	}
	return out
}

func decodeErrorResponse(t *testing.T, resp *http.Response) types.ErrorResponse {
	t.Helper()
	var out types.ErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("failed to decode ErrorResponse: %v", err)
	}
	return out
}

// --- guard tests ---

func TestHandleUploadNoProjectID(t *testing.T) {
	cfg := uploadCfg()
	cfg.ProjectID = ""
	h := HandleUpload(successWriter(), parser.Parse, cfg, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(`{}`))
	h.ServeHTTP(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
	body := decodeErrorResponse(t, resp)
	if body.Code != types.ErrInternal {
		t.Errorf("expected INTERNAL, got %q", body.Code)
	}
}

func TestHandleUploadNilFirestoreWriter(t *testing.T) {
	h := HandleUpload(nil, parser.Parse, uploadCfg(), nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(`{}`))
	h.ServeHTTP(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", resp.StatusCode)
	}
}

// --- body limit ---

func TestHandleUploadBodyTooLarge(t *testing.T) {
	cfg := uploadCfg()
	cfg.MaxFileSize = 1024
	h := HandleUpload(successWriter(), parser.Parse, cfg, nil)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(strings.Repeat("a", 2048)))
	h.ServeHTTP(w, r)
	resp := w.Result()
	if resp.StatusCode != http.StatusRequestEntityTooLarge {
		t.Errorf("expected 413, got %d", resp.StatusCode)
	}
	body := decodeErrorResponse(t, resp)
	if body.Code != types.ErrPayloadTooLarge {
		t.Errorf("expected PAYLOAD_TOO_LARGE, got %q", body.Code)
	}
}

// --- collection validation ---

func TestHandleUploadMissingCollection(t *testing.T) {
	resp := doUpload(t, successWriter(), `{"documents":[]}`)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeErrorResponse(t, resp)
	if body.Code != types.ErrCollectionRequired {
		t.Errorf("expected COLLECTION_REQUIRED, got %q", body.Code)
	}
}

func TestHandleUploadProjectMismatch(t *testing.T) {
	wrongCollection := "projects/other-project/databases/(default)/documents/sensors"
	payload := fmt.Sprintf(`{"collection":%q,"documents":[{"filename":"a.json","content":"{}","content_type":"json"}]}`, wrongCollection)
	resp := doUpload(t, successWriter(), payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeErrorResponse(t, resp)
	if body.Code != types.ErrCollectionProjectMismatch {
		t.Errorf("expected COLLECTION_PROJECT_MISMATCH, got %q", body.Code)
	}
}

func TestHandleUploadPathTraversal(t *testing.T) {
	badCollection := "projects/my-project/databases/(default)/documents/../etc"
	payload := fmt.Sprintf(`{"collection":%q,"documents":[]}`, badCollection)
	resp := doUpload(t, successWriter(), payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
}

// --- empty documents list ---

func TestHandleUploadEmptyDocuments(t *testing.T) {
	payload := fmt.Sprintf(`{"collection":%q,"documents":[]}`, validCollection)
	resp := doUpload(t, successWriter(), payload)
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	body := decodeErrorResponse(t, resp)
	if body.Code != types.ErrCollectionRequired {
		t.Errorf("expected COLLECTION_REQUIRED, got %q", body.Code)
	}
}

// --- full success (200) ---

func TestHandleUploadAllSuccess(t *testing.T) {
	payload := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"sensor.json","content":"{\"temp\":22}","content_type":"json"},
			{"filename":"note.txt","content":"hello world","content_type":"text"}
		]
	}`, validCollection)

	resp := doUpload(t, successWriter(), payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	out := decodeUploadResponse(t, resp)
	if !out.Success {
		t.Error("expected success: true")
	}
	if len(out.Uploaded) != 2 {
		t.Errorf("expected 2 uploaded, got %d", len(out.Uploaded))
	}
	if len(out.Failed) != 0 {
		t.Errorf("expected 0 failed, got %d", len(out.Failed))
	}
	for _, u := range out.Uploaded {
		if u.DocumentID == "" {
			t.Errorf("expected non-empty document_id for %s", u.Filename)
		}
		if u.Status != "created" {
			t.Errorf("expected status 'created', got %q", u.Status)
		}
	}
}

// --- partial success (207) ---

func TestHandleUploadPartialSuccess(t *testing.T) {
	payload := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"good.json","content":"{\"ok\":true}","content_type":"json"},
			{"filename":"bad.json","content":"not-json","content_type":"json"}
		]
	}`, validCollection)

	resp := doUpload(t, successWriter(), payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		t.Errorf("expected 207, got %d", resp.StatusCode)
	}
	out := decodeUploadResponse(t, resp)
	if !out.Success {
		t.Error("expected success: true on partial response")
	}
	if len(out.Uploaded) != 1 {
		t.Errorf("expected 1 uploaded, got %d", len(out.Uploaded))
	}
	if len(out.Failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(out.Failed))
	}
	if out.Failed[0].Code != types.ErrParseFailed {
		t.Errorf("expected PARSE_001, got %q", out.Failed[0].Code)
	}
}

// --- per-document error codes ---

func TestHandleUploadInvalidContentType(t *testing.T) {
	payload := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"file.csv","content":"a,b,c","content_type":"csv"}
		]
	}`, validCollection)

	resp := doUpload(t, successWriter(), payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		t.Errorf("expected 207, got %d", resp.StatusCode)
	}
	out := decodeUploadResponse(t, resp)
	if len(out.Failed) != 1 || out.Failed[0].Code != types.ErrInvalidContentType {
		t.Errorf("expected INVALID_CONTENT_TYPE, got %v", out.Failed)
	}
}

func TestHandleUploadEmptyJSON(t *testing.T) {
	payload := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"empty.json","content":"","content_type":"json"}
		]
	}`, validCollection)

	resp := doUpload(t, successWriter(), payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		t.Errorf("expected 207, got %d", resp.StatusCode)
	}
	out := decodeUploadResponse(t, resp)
	if len(out.Failed) != 1 || out.Failed[0].Code != types.ErrEmptyJSON {
		t.Errorf("expected EMPTY_JSON, got %v", out.Failed)
	}
}

func TestHandleUploadParseFailure(t *testing.T) {
	payload := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"bad.json","content":"{ invalid","content_type":"json"}
		]
	}`, validCollection)

	resp := doUpload(t, successWriter(), payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		t.Errorf("expected 207, got %d", resp.StatusCode)
	}
	out := decodeUploadResponse(t, resp)
	if len(out.Failed) != 1 || out.Failed[0].Code != types.ErrParseFailed {
		t.Errorf("expected PARSE_001, got %v", out.Failed)
	}
}

func TestHandleUploadFirestoreWriteError(t *testing.T) {
	writeErr := &firestorepkg.WriteError{
		Code:    types.ErrFirestoreQuota,
		Message: "quota exceeded",
	}
	payload := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"data.json","content":"{\"key\":\"val\"}","content_type":"json"}
		]
	}`, validCollection)

	resp := doUpload(t, errorWriter(writeErr), payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		t.Errorf("expected 207, got %d", resp.StatusCode)
	}
	out := decodeUploadResponse(t, resp)
	if len(out.Failed) != 1 || out.Failed[0].Code != types.ErrFirestoreQuota {
		t.Errorf("expected FIRESTORE_QUOTA, got %v", out.Failed)
	}
}

func TestHandleUploadFirestoreGenericError(t *testing.T) {
	payload := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"data.json","content":"{\"key\":\"val\"}","content_type":"json"}
		]
	}`, validCollection)

	resp := doUpload(t, errorWriter(errors.New("network timeout")), payload)
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		t.Errorf("expected 207, got %d", resp.StatusCode)
	}
	out := decodeUploadResponse(t, resp)
	if len(out.Failed) != 1 || out.Failed[0].Code != types.ErrInternal {
		t.Errorf("expected INTERNAL, got %v", out.Failed)
	}
}

// --- extractContentBytes ---

func TestExtractContentBytesString(t *testing.T) {
	b, err := extractContentBytes("hello")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(b) != "hello" {
		t.Errorf("expected 'hello', got %q", string(b))
	}
}

func TestExtractContentBytesBytes(t *testing.T) {
	b, err := extractContentBytes([]byte("raw"))
	if err != nil || string(b) != "raw" {
		t.Errorf("unexpected: err=%v b=%q", err, string(b))
	}
}

func TestExtractContentBytesNil(t *testing.T) {
	b, err := extractContentBytes(nil)
	if err != nil || len(b) != 0 {
		t.Errorf("unexpected: err=%v len=%d", err, len(b))
	}
}

func TestExtractContentBytesMap(t *testing.T) {
	m := map[string]any{"key": "val"}
	b, err := extractContentBytes(m)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(b, &decoded); err != nil {
		t.Fatalf("result not valid JSON: %v", err)
	}
}

// --- Shutdown closes Firestore writer ---

func TestShutdownClosesFirestoreWriter(t *testing.T) {
	closed := false
	fw := &mockWriter{
		writeFunc: func(_ context.Context, _ string, _ any) (string, error) { return "id", nil },
		closeFunc: func() error {
			closed = true
			return nil
		},
	}
	s, err := New(Config{Port: 9876}, fw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	ctx := context.Background()
	if err := s.Shutdown(ctx); err != nil {
		t.Fatalf("unexpected shutdown error: %v", err)
	}
	if !closed {
		t.Error("expected Firestore writer to be closed on Shutdown")
	}
}
