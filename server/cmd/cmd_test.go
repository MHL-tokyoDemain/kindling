package cmd

import (
	"testing"
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
