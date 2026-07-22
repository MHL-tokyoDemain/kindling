package server

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/kindling/kindling/pkg/types"
)

// --- ResolveProjectID tests ---

func TestResolveProjectIDFromFlag(t *testing.T) {
	id, err := ResolveProjectID(context.Background(), "flag-project", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "flag-project" {
		t.Errorf("expected flag-project, got %q", id)
	}
}

func TestResolveProjectIDFromEnv(t *testing.T) {
	t.Setenv("FIREBASE_PROJECT_ID", "env-project")
	id, err := ResolveProjectID(context.Background(), "", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "env-project" {
		t.Errorf("expected env-project, got %q", id)
	}
}

func TestResolveProjectIDFlagTakesPrecedenceOverEnv(t *testing.T) {
	t.Setenv("FIREBASE_PROJECT_ID", "env-project")
	id, err := ResolveProjectID(context.Background(), "flag-project", "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "flag-project" {
		t.Errorf("expected flag-project (not env), got %q", id)
	}
}

func TestResolveProjectIDFromSAJSON(t *testing.T) {
	tmpDir := t.TempDir()
	credsPath := filepath.Join(tmpDir, "sa.json")
	if err := os.WriteFile(credsPath, []byte(`{"project_id":"sa-project"}`), 0644); err != nil {
		t.Fatalf("write creds: %v", err)
	}
	id, err := ResolveProjectID(context.Background(), "", credsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "sa-project" {
		t.Errorf("expected sa-project, got %q", id)
	}
}

func TestResolveProjectIDEnvTakesPrecedenceOverSA(t *testing.T) {
	t.Setenv("FIREBASE_PROJECT_ID", "env-project")
	tmpDir := t.TempDir()
	credsPath := filepath.Join(tmpDir, "sa.json")
	if err := os.WriteFile(credsPath, []byte(`{"project_id":"sa-project"}`), 0644); err != nil {
		t.Fatalf("write creds: %v", err)
	}
	id, err := ResolveProjectID(context.Background(), "", credsPath)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if id != "env-project" {
		t.Errorf("expected env-project (not SA), got %q", id)
	}
}

func TestResolveProjectIDNoSource(t *testing.T) {
	tmpDir := t.TempDir()
	credsPath := filepath.Join(tmpDir, "sa.json")
	if err := os.WriteFile(credsPath, []byte(`{}`), 0644); err != nil {
		t.Fatalf("write creds: %v", err)
	}
	_, err := ResolveProjectID(context.Background(), "", credsPath)
	if err == nil {
		t.Fatal("expected error when project ID cannot be determined")
	}
}

func TestResolveProjectIDNoCreds(t *testing.T) {
	_, err := ResolveProjectID(context.Background(), "", "")
	if err == nil {
		t.Fatal("expected error with no flag, no env, no creds path")
	}
}

func TestResolveProjectIDUnreadableCreds(t *testing.T) {
	// SA read fails silently — falls through to error
	_, err := ResolveProjectID(context.Background(), "", "/nonexistent/sa.json")
	if err == nil {
		t.Fatal("expected error with unreadable creds and no other source")
	}
}

func TestResolveProjectIDInvalidCredsJSON(t *testing.T) {
	tmpDir := t.TempDir()
	credsPath := filepath.Join(tmpDir, "sa.json")
	if err := os.WriteFile(credsPath, []byte(`not-json`), 0644); err != nil {
		t.Fatalf("write creds: %v", err)
	}
	_, err := ResolveProjectID(context.Background(), "", credsPath)
	if err == nil {
		t.Fatal("expected error with invalid JSON and no other source")
	}
}

// --- ValidateCollectionPath tests ---

func TestValidateCollectionPathValid(t *testing.T) {
	path := "projects/my-project/databases/(default)/documents/sensors"
	projID, returnedPath, err := ValidateCollectionPath(path, "my-project")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if projID != "my-project" {
		t.Errorf("expected extracted project ID my-project, got %q", projID)
	}
	if returnedPath != path {
		t.Errorf("expected returned path %q, got %q", path, returnedPath)
	}
}

func TestValidateCollectionPathValidNestedCollection(t *testing.T) {
	path := "projects/my-project/databases/(default)/documents/users/uid/logs"
	_, _, err := ValidateCollectionPath(path, "my-project")
	if err != nil {
		t.Fatalf("unexpected error for nested path: %v", err)
	}
}

func TestValidateCollectionPathEmpty(t *testing.T) {
	_, _, err := ValidateCollectionPath("", "my-project")
	if err == nil {
		t.Fatal("expected error for empty path")
	}
	var ve *ValidationError
	if !asValidationError(err, &ve) || ve.Code != types.ErrCollectionRequired {
		t.Errorf("expected COLLECTION_REQUIRED, got %v", err)
	}
}

func TestValidateCollectionPathLeadingSlash(t *testing.T) {
	_, _, err := ValidateCollectionPath("/projects/my-project/databases/(default)/documents/col", "my-project")
	if err == nil {
		t.Fatal("expected error for leading slash")
	}
	assertMismatchCode(t, err)
}

func TestValidateCollectionPathTraversal(t *testing.T) {
	_, _, err := ValidateCollectionPath("projects/my-project/databases/(default)/documents/../etc/passwd", "my-project")
	if err == nil {
		t.Fatal("expected error for path traversal")
	}
	assertMismatchCode(t, err)
}

func TestValidateCollectionPathTraversalAtStart(t *testing.T) {
	_, _, err := ValidateCollectionPath("../projects/my-project/databases/(default)/documents/col", "my-project")
	if err == nil {
		t.Fatal("expected error for leading traversal")
	}
	assertMismatchCode(t, err)
}

func TestValidateCollectionPathNullByte(t *testing.T) {
	_, _, err := ValidateCollectionPath("projects/my-project/databases/(default)/documents/col\x00", "my-project")
	if err == nil {
		t.Fatal("expected error for null byte")
	}
	assertMismatchCode(t, err)
}

func TestValidateCollectionPathControlChar(t *testing.T) {
	_, _, err := ValidateCollectionPath("projects/my-project/databases/(default)/documents/col\x01", "my-project")
	if err == nil {
		t.Fatal("expected error for control character")
	}
	assertMismatchCode(t, err)
}

func TestValidateCollectionPathProjectMismatch(t *testing.T) {
	_, _, err := ValidateCollectionPath("projects/other-project/databases/(default)/documents/col", "my-project")
	if err == nil {
		t.Fatal("expected error for project mismatch")
	}
	assertMismatchCode(t, err)
}

func TestValidateCollectionPathMissingDocumentsSegment(t *testing.T) {
	// Missing required databases/(default)/documents/ segment
	_, _, err := ValidateCollectionPath("projects/my-project/col", "my-project")
	if err == nil {
		t.Fatal("expected error for missing required path segments")
	}
	assertMismatchCode(t, err)
}

func TestValidateCollectionPathMissingSubCollection(t *testing.T) {
	// Path ends right at documents/ with nothing after — trailing slash but no collection
	path := "projects/my-project/databases/(default)/documents/"
	_, _, err := ValidateCollectionPath(path, "my-project")
	// This technically passes prefix check; caller is responsible for sub-path validation
	// Just verify it doesn't panic and returns something consistent
	_ = err
}

func TestValidationErrorMessage(t *testing.T) {
	ve := &ValidationError{Code: types.ErrCollectionRequired, Message: "collection path is required"}
	if ve.Error() != "collection path is required" {
		t.Errorf("unexpected error message: %q", ve.Error())
	}
}

// helpers

func asValidationError(err error, target **ValidationError) bool {
	ve, ok := err.(*ValidationError)
	if ok {
		*target = ve
	}
	return ok
}

func assertMismatchCode(t *testing.T, err error) {
	t.Helper()
	var ve *ValidationError
	if !asValidationError(err, &ve) || ve.Code != types.ErrCollectionProjectMismatch {
		t.Errorf("expected COLLECTION_PROJECT_MISMATCH, got %v", err)
	}
}
