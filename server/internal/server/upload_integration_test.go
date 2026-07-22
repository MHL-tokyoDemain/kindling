//go:build integration

package server

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"cloud.google.com/go/firestore"

	"github.com/kindling/kindling/internal/parser"
)

type emulatorWriter struct {
	client    *firestore.Client
	projectID string
}

func (e *emulatorWriter) WriteDocument(ctx context.Context, collectionPath string, data any) (string, error) {
	docRef, _, err := e.client.Collection(collectionPath).Add(ctx, data)
	if err != nil {
		return "", err
	}
	return docRef.ID, nil
}

func (e *emulatorWriter) Close() error {
	return e.client.Close()
}

func (e *emulatorWriter) ProjectID() string {
	return e.projectID
}

const emulatorProject = "test-project"
const emulatorCollection = "projects/test-project/databases/(default)/documents/integration-test"

func emulatorConfig() Config {
	return Config{
		Port:        9876,
		ProjectID:   emulatorProject,
		MaxFileSize: 10 << 20,
		Concurrency: 2,
	}
}

func skipIfNoEmulator(t *testing.T) {
	t.Helper()
	if os.Getenv("FIRESTORE_EMULATOR_HOST") == "" {
		t.Skip("FIRESTORE_EMULATOR_HOST not set; skipping integration test")
	}
}

func newEmulatorClient(t *testing.T) (*firestore.Client, *emulatorWriter) {
	t.Helper()
	ctx := context.Background()
	client, err := firestore.NewClient(ctx, emulatorProject)
	if err != nil {
		t.Fatalf("failed to create emulator client: %v", err)
	}
	return client, &emulatorWriter{client: client, projectID: emulatorProject}
}

func TestEmulatorUploadJSON(t *testing.T) {
	skipIfNoEmulator(t)
	client, fw := newEmulatorClient(t)
	defer client.Close()

	handler := HandleUpload(fw, parser.Parse, emulatorConfig())

	body := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"sensor.json","content":"{\"temp\":22,\"unit\":\"c\"}","content_type":"json"}
		]
	}`, emulatorCollection)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	docs, err := client.Collection(emulatorCollection).Documents(context.Background()).GetAll()
	if err != nil {
		t.Fatalf("failed to query emulator: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("expected at least 1 document in emulator")
	}
}

func TestEmulatorUploadText(t *testing.T) {
	skipIfNoEmulator(t)
	client, fw := newEmulatorClient(t)
	defer client.Close()

	handler := HandleUpload(fw, parser.Parse, emulatorConfig())

	body := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"log.txt","content":"calibration OK","content_type":"text"}
		]
	}`, emulatorCollection)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	docs, err := client.Collection(emulatorCollection).Documents(context.Background()).GetAll()
	if err != nil {
		t.Fatalf("failed to query emulator: %v", err)
	}
	if len(docs) == 0 {
		t.Fatal("expected at least 1 document in emulator")
	}
}

func TestEmulatorInvalidJSON(t *testing.T) {
	skipIfNoEmulator(t)
	_, fw := newEmulatorClient(t)

	handler := HandleUpload(fw, parser.Parse, emulatorConfig())

	body := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"bad.json","content":"{invalid}","content_type":"json"}
		]
	}`, emulatorCollection)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusMultiStatus {
		t.Fatalf("expected 207, got %d", resp.StatusCode)
	}
}

func TestEmulatorEmptyBatch(t *testing.T) {
	skipIfNoEmulator(t)
	_, fw := newEmulatorClient(t)

	handler := HandleUpload(fw, parser.Parse, emulatorConfig())

	body := fmt.Sprintf(`{"collection": %q, "documents": []}`, emulatorCollection)

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestEmulatorProjectMismatch(t *testing.T) {
	skipIfNoEmulator(t)
	_, fw := newEmulatorClient(t)

	handler := HandleUpload(fw, parser.Parse, emulatorConfig())

	body := `{
		"collection": "projects/wrong-project/databases/(default)/documents/col",
		"documents": [
			{"filename":"a.json","content":"{}","content_type":"json"}
		]
	}`

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", resp.StatusCode)
	}
}

func TestEmulatorConcurrentBatch(t *testing.T) {
	skipIfNoEmulator(t)
	client, fw := newEmulatorClient(t)
	defer client.Close()

	handler := HandleUpload(fw, parser.Parse, emulatorConfig())

	docs := make([]string, 50)
	for i := range docs {
		docs[i] = fmt.Sprintf(`{"filename":"f%d.json","content":"{\"i\":%d}","content_type":"json"}`, i, i)
	}

	body := fmt.Sprintf(`{"collection":%q,"documents":[%s]}`, emulatorCollection, strings.Join(docs, ","))

	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(body))
	r.Header.Set("Content-Type", "application/json")
	handler.ServeHTTP(w, r)

	resp := w.Result()
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}

	snapshot, err := client.Collection(emulatorCollection).Documents(context.Background()).GetAll()
	if err != nil {
		t.Fatalf("failed to query emulator: %v", err)
	}
	if len(snapshot) != 50 {
		t.Errorf("expected 50 documents, got %d", len(snapshot))
	}
}
