package server

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

const (
	AuditLevelINFO  = "INFO"
	AuditLevelWARN  = "WARN"
	AuditLevelERROR = "ERROR"
)

type AuditEntry struct {
	Timestamp      string `json:"timestamp"`
	TransactionID  string `json:"transaction_id"`
	Level          string `json:"level"`
	Filename       string `json:"filename"`
	CollectionPath string `json:"collection_path"`
	Outcome        string `json:"outcome"`
}

type AuditLogger struct {
	mu     sync.Mutex
	file   *os.File
	enc    *json.Encoder
	closed bool
}

func NewAuditLogger(filePath string) (*AuditLogger, error) {
	f, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open audit log: %w", err)
	}
	return &AuditLogger{
		file: f,
		enc:  json.NewEncoder(f),
	}, nil
}

func (l *AuditLogger) Log(level, transactionID, filename, collectionPath, outcome string) {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return
	}
	l.enc.Encode(AuditEntry{
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		TransactionID:  transactionID,
		Level:          level,
		Filename:       filename,
		CollectionPath: collectionPath,
		Outcome:        outcome,
	})
}

func (l *AuditLogger) Close() error {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.closed {
		return nil
	}
	l.closed = true
	return l.file.Close()
}

func TransactionID() string {
	uuid := make([]byte, 16)
	_, _ = rand.Read(uuid)
	uuid[6] = (uuid[6] & 0x0f) | 0x40
	uuid[8] = (uuid[8] & 0x3f) | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", uuid[0:4], uuid[4:6], uuid[6:8], uuid[8:10], uuid[10:])
}
