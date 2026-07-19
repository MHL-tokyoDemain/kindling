package main

import (
	"os"
	"testing"
)

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
	code := run([]string{"kindling", "server", "-port", "8080"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}

func TestRunUploadDispatch(t *testing.T) {
	dir := t.TempDir()
	f1 := dir + "/a.json"
	f2 := dir + "/b.txt"
	if err := os.WriteFile(f1, []byte(`{"x":1}`), 0644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("hello"), 0644); err != nil {
		t.Fatal(err)
	}

	code := run([]string{"kindling", "upload", "-collection", "test-col", "-files", f1 + "," + f2, "-max-file-size", "1048576"})
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
}
