package firestore

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/kindling/kindling/pkg/types"
)

type mockFirestoreWriter struct {
	writeDocumentFunc func(ctx context.Context, collectionPath string, data any) (string, error)
	closeFunc         func() error
	projectIDFunc     func() string
}

func (m *mockFirestoreWriter) WriteDocument(ctx context.Context, collectionPath string, data any) (string, error) {
	return m.writeDocumentFunc(ctx, collectionPath, data)
}

func (m *mockFirestoreWriter) Close() error {
	return m.closeFunc()
}

func (m *mockFirestoreWriter) ProjectID() string {
	return m.projectIDFunc()
}

func TestFirestoreWriterInterface(t *testing.T) {
	var _ FirestoreWriter = (*Client)(nil)
}

func TestMockFirestoreWriter(t *testing.T) {
	mock := &mockFirestoreWriter{
		writeDocumentFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			return "doc-123", nil
		},
		closeFunc: func() error {
			return nil
		},
		projectIDFunc: func() string {
			return "test-project"
		},
	}

	docID, err := mock.WriteDocument(context.Background(), "test/docs", map[string]string{"key": "value"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if docID != "doc-123" {
		t.Errorf("expected doc-123, got %q", docID)
	}

	if mock.ProjectID() != "test-project" {
		t.Errorf("expected test-project, got %q", mock.ProjectID())
	}

	if err := mock.Close(); err != nil {
		t.Errorf("unexpected close error: %v", err)
	}
}

func TestResolveCredentials(t *testing.T) {
	tmpDir := t.TempDir()
	credsPath := filepath.Join(tmpDir, "serviceAccountKey.json")
	if err := os.WriteFile(credsPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write test creds: %v", err)
	}

	path, err := resolveCredentials(credsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != credsPath {
		t.Errorf("expected %q, got %q", credsPath, path)
	}
}

func TestResolveCredentialsEnvVar(t *testing.T) {
	tmpDir := t.TempDir()
	credsPath := filepath.Join(tmpDir, "env-creds.json")
	if err := os.WriteFile(credsPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write test creds: %v", err)
	}

	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", credsPath)
	path, err := resolveCredentials("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != credsPath {
		t.Errorf("expected %q, got %q", credsPath, path)
	}
}

func TestResolveCredentialsDefaultPath(t *testing.T) {
	origDir, _ := os.Getwd()
	tmpDir := t.TempDir()
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir)

	if err := os.WriteFile("serviceAccountKey.json", []byte("{}"), 0644); err != nil {
		t.Fatalf("failed to write test creds: %v", err)
	}

	path, err := resolveCredentials("")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if path != "./serviceAccountKey.json" {
		t.Errorf("expected ./serviceAccountKey.json, got %q", path)
	}
}

func TestResolveCredentialsNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to getwd: %v", err)
	}
	if err := os.Chdir(tmpDir); err != nil {
		t.Fatalf("failed to chdir: %v", err)
	}
	defer os.Chdir(origDir)

	_, err = resolveCredentials("/nonexistent/path.json")
	if err == nil {
		t.Fatal("expected error for missing credentials")
	}
}

func TestResolveProjectIDOverride(t *testing.T) {
	id, err := resolveProjectID("my-project", "/ignored/path.json")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "my-project" {
		t.Errorf("expected my-project, got %q", id)
	}
}

func TestResolveProjectIDFromFile(t *testing.T) {
	tmpDir := t.TempDir()
	credsPath := filepath.Join(tmpDir, "creds.json")
	credsContent := `{"project_id": "from-file-project"}`
	if err := os.WriteFile(credsPath, []byte(credsContent), 0644); err != nil {
		t.Fatalf("failed to write creds: %v", err)
	}

	id, err := resolveProjectID("", credsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "from-file-project" {
		t.Errorf("expected from-file-project, got %q", id)
	}
}

func TestResolveProjectIDEmpty(t *testing.T) {
	tmpDir := t.TempDir()
	credsPath := filepath.Join(tmpDir, "creds.json")
	if err := os.WriteFile(credsPath, []byte(`{}`), 0644); err != nil {
		t.Fatalf("failed to write creds: %v", err)
	}

	_, err := resolveProjectID("", credsPath)
	if !errors.Is(err, ErrProjectIDRequired) {
		t.Errorf("expected ErrProjectIDRequired, got %v", err)
	}
}

func TestResolveProjectIDUnreadableFile(t *testing.T) {
	_, err := resolveProjectID("", "/nonexistent/creds.json")
	if err == nil {
		t.Fatal("expected error for unreadable file")
	}
}

func TestWriteErrorCode(t *testing.T) {
	err := &WriteError{
		Code:    types.ErrFirestoreQuota,
		Message: "quota exceeded",
	}

	if err.Error() != "quota exceeded" {
		t.Errorf("expected 'quota exceeded', got %q", err.Error())
	}
	if err.Code != types.ErrFirestoreQuota {
		t.Errorf("expected code %q, got %q", types.ErrFirestoreQuota, err.Code)
	}
}

func TestWriteErrorUnwrap(t *testing.T) {
	inner := errors.New("inner error")
	err := &WriteError{
		Code:    types.ErrInternal,
		Message: "wrapped error",
		Err:     inner,
	}

	if !errors.Is(err, inner) {
		t.Error("expected WriteError to wrap inner error")
	}
}

func TestMockWriteDocumentError(t *testing.T) {
	expectedErr := errors.New("write failed")
	mock := &mockFirestoreWriter{
		writeDocumentFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			return "", expectedErr
		},
	}

	_, err := mock.WriteDocument(context.Background(), "test/docs", nil)
	if !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}

func TestMockCloseError(t *testing.T) {
	expectedErr := errors.New("close failed")
	mock := &mockFirestoreWriter{
		writeDocumentFunc: func(ctx context.Context, collectionPath string, data any) (string, error) {
			return "id", nil
		},
		closeFunc: func() error {
			return expectedErr
		},
	}

	if err := mock.Close(); !errors.Is(err, expectedErr) {
		t.Errorf("expected %v, got %v", expectedErr, err)
	}
}
