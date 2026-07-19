package cmd

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kindling/kindling/pkg/types"
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

func TestRunUploadFileReadError(t *testing.T) {
	code, err := RunUpload("my-collection", []string{"/nonexistent/file.json"}, "", "", 1048576, 0)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if err != nil {
		t.Errorf("expected no error from RunUpload, got %v", err)
	}
}

func TestRunUploadParseError(t *testing.T) {
	dir := t.TempDir()
	badFile := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badFile, []byte(`{invalid}`), 0644); err != nil {
		t.Fatal(err)
	}

	code, err := RunUpload("my-collection", []string{badFile}, "", "", 1048576, 0)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if err != nil {
		t.Errorf("expected no error from RunUpload, got %v", err)
	}
}

func TestRunUploadFullSuccess(t *testing.T) {
	dir := t.TempDir()

	jsonFile := filepath.Join(dir, "data.json")
	if err := os.WriteFile(jsonFile, []byte(`{"name": "test"}`), 0644); err != nil {
		t.Fatal(err)
	}

	textFile := filepath.Join(dir, "log.txt")
	if err := os.WriteFile(textFile, []byte("hello world"), 0644); err != nil {
		t.Fatal(err)
	}

	code, err := RunUpload("my-collection", []string{jsonFile, textFile}, "", "", 1048576, 0)
	if code != 0 {
		t.Errorf("expected exit code 0, got %d", code)
	}
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestRunUploadMixedSuccessFailure(t *testing.T) {
	dir := t.TempDir()

	goodFile := filepath.Join(dir, "good.json")
	if err := os.WriteFile(goodFile, []byte(`{"ok": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	badFile := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badFile, []byte(`{invalid}`), 0644); err != nil {
		t.Fatal(err)
	}

	missingFile := filepath.Join(dir, "missing.txt")

	code, err := RunUpload("my-collection", []string{goodFile, badFile, missingFile}, "", "", 1048576, 0)
	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if err != nil {
		t.Errorf("expected no error from top level, got %v", err)
	}
}

func TestRunUploadOutputJSON(t *testing.T) {
	dir := t.TempDir()

	goodFile := filepath.Join(dir, "good.json")
	if err := os.WriteFile(goodFile, []byte(`{"ok": true}`), 0644); err != nil {
		t.Fatal(err)
	}

	badFile := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(badFile, []byte(`{invalid}`), 0644); err != nil {
		t.Fatal(err)
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	old := os.Stdout
	os.Stdout = w

	code, topErr := RunUpload("my-collection", []string{goodFile, badFile}, "", "", 1048576, 0)

	w.Close()
	os.Stdout = old

	var outBytes []byte
	outBytes = make([]byte, 4096)
	n, _ := r.Read(outBytes)
	outBytes = outBytes[:n]

	if code != 1 {
		t.Errorf("expected exit code 1, got %d", code)
	}
	if topErr != nil {
		t.Errorf("expected no top-level error, got %v", topErr)
	}

	var resp types.UploadResponse
	if err := json.Unmarshal(outBytes, &resp); err != nil {
		t.Fatalf("failed to unmarshal output: %v\noutput: %s", err, string(outBytes))
	}
	if resp.Collection != "my-collection" {
		t.Errorf("expected collection 'my-collection', got %q", resp.Collection)
	}
	if resp.Success {
		t.Error("expected success=false with mixed results")
	}
	if len(resp.Uploaded) != 1 {
		t.Errorf("expected 1 uploaded, got %d", len(resp.Uploaded))
	}
	if len(resp.Failed) != 1 {
		t.Errorf("expected 1 failed, got %d", len(resp.Failed))
	}
	if resp.Uploaded[0].Filename != goodFile {
		t.Errorf("expected uploaded file %q, got %q", goodFile, resp.Uploaded[0].Filename)
	}
	if resp.Failed[0].Filename != badFile {
		t.Errorf("expected failed file %q, got %q", badFile, resp.Failed[0].Filename)
	}
}
