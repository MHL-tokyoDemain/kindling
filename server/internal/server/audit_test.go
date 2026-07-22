package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/kindling/kindling/internal/parser"
)

func TestAuditLoggerWritesJSONLines(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}
	defer l.Close()

	tid := "test-txn-1"
	l.Log(AuditLevelINFO, tid, "f1.json", "col/a", "doc-abc")

	lines := readAuditLog(t, path)
	if len(lines) != 1 {
		t.Fatalf("expected 1 line, got %d", len(lines))
	}
	var entry AuditEntry
	if err := json.Unmarshal([]byte(lines[0]), &entry); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if entry.TransactionID != tid {
		t.Errorf("transaction_id: got %q, want %q", entry.TransactionID, tid)
	}
	if entry.Level != AuditLevelINFO {
		t.Errorf("level: got %q, want %q", entry.Level, AuditLevelINFO)
	}
	if entry.Filename != "f1.json" {
		t.Errorf("filename: got %q, want %q", entry.Filename, "f1.json")
	}
	if entry.CollectionPath != "col/a" {
		t.Errorf("collection_path: got %q, want %q", entry.CollectionPath, "col/a")
	}
	if entry.Outcome != "doc-abc" {
		t.Errorf("outcome: got %q, want %q", entry.Outcome, "doc-abc")
	}
	if entry.Timestamp == "" {
		t.Error("timestamp is empty")
	}
}

func TestAuditLoggerMultipleEntries(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}
	defer l.Close()

	l.Log(AuditLevelINFO, "t1", "a.json", "col/a", "doc-1")
	l.Log(AuditLevelERROR, "t1", "b.json", "col/a", "PARSE_001")
	l.Log(AuditLevelINFO, "t2", "c.json", "col/b", "doc-2")

	lines := readAuditLog(t, path)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	var e2 AuditEntry
	json.Unmarshal([]byte(lines[1]), &e2)
	if e2.Filename != "b.json" || e2.Outcome != "PARSE_001" {
		t.Errorf("second entry mismatch: %+v", e2)
	}
}

func TestAuditLoggerAppend(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "audit.log")

	l1, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}
	l1.Log(AuditLevelINFO, "t1", "a.json", "col/a", "doc-1")
	l1.Close()

	l2, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger (append): %v", err)
	}
	l2.Log(AuditLevelWARN, "t1", "b.json", "col/a", "QUOTA")
	l2.Close()

	lines := readAuditLog(t, path)
	if len(lines) != 2 {
		t.Fatalf("expected 2 lines, got %d", len(lines))
	}
}

func TestAuditLoggerLevels(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}
	defer l.Close()

	levels := []string{AuditLevelINFO, AuditLevelWARN, AuditLevelERROR}
	for i, lvl := range levels {
		l.Log(lvl, "tid", fmt.Sprintf("f%d.json", i), "col", lvl)
	}

	lines := readAuditLog(t, path)
	if len(lines) != 3 {
		t.Fatalf("expected 3 lines, got %d", len(lines))
	}
	for i, lvl := range levels {
		var e AuditEntry
		json.Unmarshal([]byte(lines[i]), &e)
		if e.Level != lvl {
			t.Errorf("line %d: expected level %q, got %q", i, lvl, e.Level)
		}
	}
}

func TestAuditLoggerNoPII(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}
	defer l.Close()

	l.Log(AuditLevelINFO, "tid", "f.json", "col/a", "doc-abc")

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	var rawMap map[string]any
	if err := json.Unmarshal(raw, &rawMap); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	piiKeys := []string{"email", "user", "password", "secret", "token", "ssn", "phone", "address"}
	for _, k := range piiKeys {
		if _, exists := rawMap[k]; exists {
			t.Errorf("unexpected PII key found: %q", k)
		}
	}
}

func TestAuditLoggerConcurrentWrites(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}
	defer l.Close()

	n := 50
	var wg sync.WaitGroup
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			l.Log(AuditLevelINFO, "t", fmt.Sprintf("f%d.json", i), "col", fmt.Sprintf("doc-%d", i))
		}(i)
	}
	wg.Wait()

	lines := readAuditLog(t, path)
	if len(lines) != n {
		t.Errorf("expected %d lines, got %d", n, len(lines))
	}
}

func TestTransactionIDFormat(t *testing.T) {
	tid := TransactionID()
	if len(tid) != 36 {
		t.Errorf("expected 36 chars, got %d: %q", len(tid), tid)
	}
	for i, c := range tid {
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				t.Errorf("expected '-' at pos %d, got %c", i, c)
			}
		}
	}
}

func TestTransactionIDUniqueness(t *testing.T) {
	n := 100
	seen := make(map[string]bool)
	for i := 0; i < n; i++ {
		tid := TransactionID()
		if seen[tid] {
			t.Fatalf("duplicate transaction_id: %s", tid)
		}
		seen[tid] = true
	}
}

func TestCloseIdempotent(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Errorf("first close: %v", err)
	}
	if err := l.Close(); err != nil {
		t.Errorf("second close: %v", err)
	}
}

func TestLogAfterCloseIsNoop(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	l, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}
	l.Close()
	l.Log(AuditLevelINFO, "tid", "f.json", "col", "doc")

	lines := readAuditLog(t, path)
	if len(lines) != 0 {
		t.Errorf("expected 0 lines after close, got %d", len(lines))
	}
}

func TestHandleUploadWithAuditLogger(t *testing.T) {
	path := filepath.Join(t.TempDir(), "audit.log")
	al, err := NewAuditLogger(path)
	if err != nil {
		t.Fatalf("NewAuditLogger: %v", err)
	}
	defer al.Close()

	payload := fmt.Sprintf(`{
		"collection": %q,
		"documents": [
			{"filename":"a.json","content":"{\"k\":\"v\"}","content_type":"json"},
			{"filename":"b.json","content":"not-json","content_type":"json"}
		]
	}`, validCollection)

	h := HandleUpload(successWriter(), parser.Parse, uploadCfg(), al)
	w := httptest.NewRecorder()
	r := httptest.NewRequest("POST", "/upload", strings.NewReader(payload))
	h.ServeHTTP(w, r)

	lines := readAuditLog(t, path)
	if len(lines) != 2 {
		t.Fatalf("expected 2 audit entries, got %d", len(lines))
	}

	var entries []AuditEntry
	for _, line := range lines {
		var e AuditEntry
		json.Unmarshal([]byte(line), &e)
		entries = append(entries, e)
	}

	if entries[0].TransactionID != entries[1].TransactionID {
		t.Error("entries in same batch should share transaction_id")
	}
	if entries[0].TransactionID == "" {
		t.Error("transaction_id should not be empty")
	}

	var foundInfo, foundError bool
	for _, e := range entries {
		switch e.Level {
		case AuditLevelINFO:
			foundInfo = true
			if e.Filename != "a.json" {
				t.Errorf("info entry filename: got %q, want %q", e.Filename, "a.json")
			}
		case AuditLevelERROR:
			foundError = true
			if e.Filename != "b.json" {
				t.Errorf("error entry filename: got %q, want %q", e.Filename, "b.json")
			}
			if e.Outcome != "PARSE_001" {
				t.Errorf("error entry outcome: got %q, want %q", e.Outcome, "PARSE_001")
			}
		default:
			t.Errorf("unexpected level: %q", e.Level)
		}
	}
	if !foundInfo {
		t.Error("missing INFO entry")
	}
	if !foundError {
		t.Error("missing ERROR entry")
	}
}

func TestNewAuditLoggerBadPath(t *testing.T) {
	_, err := NewAuditLogger("/nonexistent/deeply/nested/audit.log")
	if err == nil {
		t.Error("expected error for bad path")
	}
}

func readAuditLog(t *testing.T, path string) []string {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer f.Close()
	var lines []string
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		t.Fatalf("scanner error: %v", err)
	}
	return lines
}
