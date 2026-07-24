package cmd

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kindling/kindling/internal/firestore"
)

func TestRunUploadMissingCollection(t *testing.T) {
	code, err := RunUpload("", nil, "", "", 0, 0)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "--collection is required" {
		t.Errorf("expected '--collection is required', got %q", err.Error())
	}
}

func TestRunUploadMissingFiles(t *testing.T) {
	code, err := RunUpload("my-collection", nil, "", "", 0, 0)
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
	if err == nil {
		t.Fatal("expected error, got nil")
	}
	if err.Error() != "--files is required" {
		t.Errorf("expected '--files is required', got %q", err.Error())
	}
}

// mockWriter implements firestore.FirestoreWriter for testing uploadFiles.
type mockWriter struct {
	writeFunc func(ctx context.Context, collectionPath string, data any) (string, error)
	closed    bool
}

func (m *mockWriter) WriteDocument(ctx context.Context, collectionPath string, data any) (string, error) {
	return m.writeFunc(ctx, collectionPath, data)
}

func (m *mockWriter) Close() error {
	m.closed = true
	return nil
}

func (m *mockWriter) ProjectID() string {
	return "test-project"
}

func writeTempFile(t *testing.T, name, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write temp file: %v", err)
	}
	return path
}

func TestUploadFilesSuccess(t *testing.T) {
	jsonFile := writeTempFile(t, "data.json", `{"name":"test"}`)
	textFile := writeTempFile(t, "log.txt", "hello world")

	mock := &mockWriter{
		writeFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			return "doc-abc", nil
		},
	}

	code, err := uploadFiles(context.Background(), mock, "my-collection", []string{jsonFile, textFile}, 1048576)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestUploadFilesFileReadError(t *testing.T) {
	mock := &mockWriter{
		writeFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			return "doc-abc", nil
		},
	}

	code, err := uploadFiles(context.Background(), mock, "my-collection", []string{"/nonexistent/file.json"}, 1048576)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if code != 1 {
		t.Errorf("expected exit code 1 for read failure, got %d", code)
	}
}

func TestUploadFilesParseError(t *testing.T) {
	badFile := writeTempFile(t, "bad.json", `{invalid json`)

	mock := &mockWriter{
		writeFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			return "doc-abc", nil
		},
	}

	code, err := uploadFiles(context.Background(), mock, "my-collection", []string{badFile}, 1048576)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if code != 1 {
		t.Errorf("expected exit code 1 for parse failure, got %d", code)
	}
}

func TestUploadFilesWriteErrorWithCode(t *testing.T) {
	jsonFile := writeTempFile(t, "data.json", `{"name":"test"}`)

	mock := &mockWriter{
		writeFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			return "", &firestore.WriteError{
				Code:    "FIRESTORE_QUOTA",
				Message: "quota exceeded",
			}
		},
	}

	code, err := uploadFiles(context.Background(), mock, "my-collection", []string{jsonFile}, 1048576)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if code != 1 {
		t.Errorf("expected exit code 1 for write failure, got %d", code)
	}
}

func TestUploadFilesWriteErrorPlainError(t *testing.T) {
	jsonFile := writeTempFile(t, "data.json", `{"name":"test"}`)

	mock := &mockWriter{
		writeFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			return "", errors.New("generic write failure")
		},
	}

	code, err := uploadFiles(context.Background(), mock, "my-collection", []string{jsonFile}, 1048576)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
}

func TestUploadFilesMixed(t *testing.T) {
	goodFile := writeTempFile(t, "good.json", `{"ok":true}`)
	badFile := writeTempFile(t, "bad.json", `{invalid`)

	mock := &mockWriter{
		writeFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			return "doc-good", nil
		},
	}

	code, err := uploadFiles(context.Background(), mock, "my-collection", []string{goodFile, badFile}, 1048576)
	if err != nil {
		t.Fatalf("unexpected top-level error: %v", err)
	}
	if code != 1 {
		t.Errorf("expected exit code 1 with one failure, got %d", code)
	}
}

func TestUploadFilesAllSuccess(t *testing.T) {
	f1 := writeTempFile(t, "a.json", `{"a":1}`)
	f2 := writeTempFile(t, "b.json", `{"b":2}`)

	var writeCount int
	mock := &mockWriter{
		writeFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			writeCount++
			return "doc", nil
		},
	}

	code, err := uploadFiles(context.Background(), mock, "my-collection", []string{f1, f2}, 1048576)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if writeCount != 2 {
		t.Errorf("expected 2 writes, got %d", writeCount)
	}
}
