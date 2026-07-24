package main

import (
	"os"
	"path/filepath"
	"testing"
)

// isolateCredentials points credential discovery at an empty temp dir so that
// Firestore client creation fails deterministically without external state.
func isolateCredentials(t *testing.T) {
	t.Helper()
	t.Setenv("GOOGLE_APPLICATION_CREDENTIALS", "")
	origDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if err := os.Chdir(t.TempDir()); err != nil {
		t.Fatalf("chdir: %v", err)
	}
	t.Cleanup(func() { _ = os.Chdir(origDir) })
}

func TestRunNoArgs(t *testing.T) {
	code := run([]string{"kindling"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestRunInvalidCommand(t *testing.T) {
	code := run([]string{"kindling", "bogus"})
	if code != 2 {
		t.Errorf("expected exit code 2, got %d", code)
	}
}

func TestEnvStringSet(t *testing.T) {
	t.Setenv("KINDLING_TEST_ENV_STRING", "custom")
	got := envString("KINDLING_TEST_ENV_STRING", "default")
	if got != "custom" {
		t.Errorf("expected 'custom', got %q", got)
	}
}

func TestEnvStringUnset(t *testing.T) {
	got := envString("KINDLING_TEST_ENV_NONEXISTENT", "fallback")
	if got != "fallback" {
		t.Errorf("expected 'fallback', got %q", got)
	}
}

func TestEnvIntValid(t *testing.T) {
	t.Setenv("KINDLING_TEST_ENV_INT", "42")
	got := envInt("KINDLING_TEST_ENV_INT", 0)
	if got != 42 {
		t.Errorf("expected 42, got %d", got)
	}
}

func TestEnvIntInvalid(t *testing.T) {
	t.Setenv("KINDLING_TEST_ENV_INT", "not-a-number")
	got := envInt("KINDLING_TEST_ENV_INT", 99)
	if got != 99 {
		t.Errorf("expected fallback 99, got %d", got)
	}
}

func TestEnvIntUnset(t *testing.T) {
	got := envInt("KINDLING_TEST_ENV_NONEXISTENT", 77)
	if got != 77 {
		t.Errorf("expected fallback 77, got %d", got)
	}
}

func TestEnvInt64Valid(t *testing.T) {
	t.Setenv("KINDLING_TEST_ENV_INT64", "1048576")
	got := envInt64("KINDLING_TEST_ENV_INT64", 0)
	if got != 1048576 {
		t.Errorf("expected 1048576, got %d", got)
	}
}

func TestEnvInt64Invalid(t *testing.T) {
	t.Setenv("KINDLING_TEST_ENV_INT64", "bogus")
	got := envInt64("KINDLING_TEST_ENV_INT64", 512)
	if got != 512 {
		t.Errorf("expected fallback 512, got %d", got)
	}
}

func TestEnvInt64Unset(t *testing.T) {
	got := envInt64("KINDLING_TEST_ENV_NONEXISTENT", 256)
	if got != 256 {
		t.Errorf("expected fallback 256, got %d", got)
	}
}

func TestEnvStringEmpty(t *testing.T) {
	t.Setenv("KINDLING_TEST_ENV_EMPTY", "")
	got := envString("KINDLING_TEST_ENV_EMPTY", "default")
	if got != "default" {
		t.Errorf("expected 'default' for empty env var, got %q", got)
	}
}

func TestRunServerDispatch(t *testing.T) {
	isolateCredentials(t)
	// No credentials available, so RunServer fails during project resolution.
	// This still exercises the server dispatch path in run().
	code := run([]string{"kindling", "server", "-port", "9877", "-creds", "/nonexistent/creds.json"})
	if code != 1 {
		t.Errorf("expected exit code 1 on server startup failure, got %d", code)
	}
}

func TestRunServerBadFlag(t *testing.T) {
	code := run([]string{"kindling", "server", "-nonexistent-flag"})
	if code != 2 {
		t.Errorf("expected exit code 2 on bad flag, got %d", code)
	}
}

func TestRunUploadDispatch(t *testing.T) {
	isolateCredentials(t)
	dir := t.TempDir()
	f1 := filepath.Join(dir, "a.json")
	if err := os.WriteFile(f1, []byte(`{"x":1}`), 0644); err != nil {
		t.Fatal(err)
	}

	// No credentials available, so RunUpload fails at Firestore client creation
	// and returns exit code 3. This exercises the upload dispatch path in run().
	code := run([]string{"kindling", "upload", "-collection", "test-col", "-files", f1, "-creds", "/nonexistent/creds.json"})
	if code != 3 {
		t.Errorf("expected exit code 3 on Firestore client failure, got %d", code)
	}
}

func TestRunUploadBadFlag(t *testing.T) {
	code := run([]string{"kindling", "upload", "-nonexistent-flag"})
	if code != 2 {
		t.Errorf("expected exit code 2 on bad flag, got %d", code)
	}
}

func TestRunUploadDispatchMissingCollection(t *testing.T) {
	// Upload dispatch with no collection flag returns validation exit code 2.
	code := run([]string{"kindling", "upload", "-files", "/tmp/whatever.json"})
	if code != 2 {
		t.Errorf("expected exit code 2 for missing collection, got %d", code)
	}
}
