package server

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"unicode"

	"github.com/kindling/kindling/pkg/types"
)

// ValidationError carries an error code (from types constants) alongside a human-readable message.
type ValidationError struct {
	Code    string
	Message string
}

func (e *ValidationError) Error() string {
	return e.Message
}

// ResolveProjectID determines the Firebase project ID using the precedence order:
//  1. projectFlag (--project CLI flag; main.go also populates this from FIREBASE_PROJECT_ID env var)
//  2. FIREBASE_PROJECT_ID environment variable (fallback if called without main.go's env mapping)
//  3. project_id field inside the service-account JSON at credsPath
//
// Returns an error if the project ID cannot be determined by any of the above steps.
// SDK auto-detection (step 4) is not performed here to keep this function unit-testable.
func ResolveProjectID(_ context.Context, projectFlag string, credsPath string) (string, error) {
	// Step 1: explicit flag value (may already include env var resolved by main.go)
	if projectFlag != "" {
		return projectFlag, nil
	}

	// Step 2: FIREBASE_PROJECT_ID environment variable
	if v := os.Getenv("FIREBASE_PROJECT_ID"); v != "" {
		return v, nil
	}

	// Step 3: project_id field in service-account JSON
	if credsPath != "" {
		if id, err := readProjectIDFromSA(credsPath); err == nil && id != "" {
			return id, nil
		}
	}

	return "", fmt.Errorf(
		"project ID could not be determined: provide --project, set FIREBASE_PROJECT_ID, " +
			"or ensure the service-account JSON contains a project_id field",
	)
}

// readProjectIDFromSA extracts the project_id field from a service-account JSON file.
func readProjectIDFromSA(credsPath string) (string, error) {
	data, err := os.ReadFile(credsPath)
	if err != nil {
		return "", fmt.Errorf("cannot read credentials file: %w", err)
	}
	var sa struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(data, &sa); err != nil {
		return "", fmt.Errorf("invalid credentials JSON: %w", err)
	}
	return sa.ProjectID, nil
}

// ValidateCollectionPath checks that collectionPath is a well-formed Firestore resource path
// that is bound to boundProjectID.
//
// A valid path must:
//   - Be non-empty
//   - Not start with "/"
//   - Not contain ".." segments (path traversal)
//   - Not contain null bytes or other control characters
//   - Match the prefix: projects/<boundProjectID>/databases/(default)/documents/
//
// Returns the project ID extracted from the path, the original collectionPath, and nil on success.
// Returns a *ValidationError with the appropriate code on failure.
func ValidateCollectionPath(collectionPath, boundProjectID string) (string, string, error) {
	// Empty collection
	if collectionPath == "" {
		return "", "", &ValidationError{
			Code:    types.ErrCollectionRequired,
			Message: "collection path is required",
		}
	}

	// Leading slash
	if strings.HasPrefix(collectionPath, "/") {
		return "", "", &ValidationError{
			Code:    types.ErrCollectionProjectMismatch,
			Message: "collection path must not start with '/'",
		}
	}

	// Null bytes and control characters
	for _, r := range collectionPath {
		if r == 0 || (r != '\t' && unicode.IsControl(r)) {
			return "", "", &ValidationError{
				Code:    types.ErrCollectionProjectMismatch,
				Message: "collection path contains invalid characters",
			}
		}
	}

	// Path traversal
	for _, segment := range strings.Split(collectionPath, "/") {
		if segment == ".." {
			return "", "", &ValidationError{
				Code:    types.ErrCollectionProjectMismatch,
				Message: "collection path must not contain '..' segments",
			}
		}
	}

	// Must start with projects/<boundProjectID>/databases/(default)/documents/
	expectedPrefix := fmt.Sprintf("projects/%s/databases/(default)/documents/", boundProjectID)
	if !strings.HasPrefix(collectionPath, expectedPrefix) {
		return "", "", &ValidationError{
			Code:    types.ErrCollectionProjectMismatch,
			Message: fmt.Sprintf("collection path must start with %q", expectedPrefix),
		}
	}

	// Extract project ID from path for caller confirmation
	// Format: projects/<projectID>/databases/...
	parts := strings.SplitN(collectionPath, "/", 3)
	extractedProjectID := parts[1]

	return extractedProjectID, collectionPath, nil
}
