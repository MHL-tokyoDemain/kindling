package server

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"sync"

	firestorepkg "github.com/kindling/kindling/internal/firestore"
	"github.com/kindling/kindling/internal/parser"
	"github.com/kindling/kindling/pkg/types"
)

// firestoreWriter is the interface the upload handler uses to write documents.
// It mirrors firestore.FirestoreWriter to allow mock injection in tests.
type firestoreWriter interface {
	WriteDocument(ctx context.Context, collectionPath string, data any) (string, error)
	Close() error
	ProjectID() string
}

// parserFunc is the signature of parser.Parse, allowing injection in tests.
type parserFunc func(filename string, content []byte, maxFileSize int64) parser.ParseResult

// docResult holds the outcome of processing a single document.
type docResult struct {
	uploaded *types.UploadResult
	failed   *types.UploadResult
}

// HandleUpload returns an http.HandlerFunc that processes POST /upload requests.
// It validates the request, processes each document concurrently up to cfg.Concurrency
// goroutines, writes successful documents to Firestore, and returns a 200 (all success)
// or 207 (partial success) response.
func HandleUpload(fw firestoreWriter, parse parserFunc, cfg Config, auditLogger *AuditLogger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// 1. Enforce body size limit — checked first regardless of other configuration.
		r.Body = http.MaxBytesReader(w, r.Body, cfg.MaxFileSize)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				WriteError(w, types.ErrPayloadTooLarge, "request body exceeds size limit")
			} else {
				WriteError(w, types.ErrInternal, "failed to read request body")
			}
			return
		}

		// Guard: server must be bound to a project.
		if cfg.ProjectID == "" {
			WriteError(w, types.ErrInternal, "server is not bound to a project ID")
			return
		}

		// Guard: Firestore client must be configured.
		if fw == nil {
			WriteError(w, types.ErrInternal, "Firestore client is not configured")
			return
		}

		// 2. Decode request JSON.
		var req types.UploadRequest
		if err := json.Unmarshal(body, &req); err != nil {
			WriteError(w, types.ErrInternal, "invalid JSON in request body")
			return
		}

		// 3. Validate collection path.
		if req.Collection == "" {
			WriteError(w, types.ErrCollectionRequired, "collection path is required")
			return
		}
		if _, _, valErr := ValidateCollectionPath(req.Collection, cfg.ProjectID); valErr != nil {
			var ve *ValidationError
			if errors.As(valErr, &ve) {
				WriteError(w, ve.Code, ve.Message)
			} else {
				WriteError(w, types.ErrInternal, "collection path validation failed")
			}
			return
		}

		// 4. Require at least one document.
		if len(req.Documents) == 0 {
			WriteError(w, types.ErrCollectionRequired, "at least one document is required")
			return
		}

		// 5. Generate transaction ID for audit logging.
		transactionID := TransactionID()

		// 6. Process documents concurrently.
		concurrency := cfg.Concurrency
		if concurrency <= 0 {
			concurrency = 10
		}

		results := make([]docResult, len(req.Documents))
		sem := make(chan struct{}, concurrency)
		var wg sync.WaitGroup

		for i, doc := range req.Documents {
			wg.Add(1)
			sem <- struct{}{}
			go func(idx int, doc types.Document) {
				defer wg.Done()
				defer func() { <-sem }()
				res := processDocument(r.Context(), doc, req.Collection, parse, fw, cfg.MaxFileSize)
				results[idx] = res
				if auditLogger != nil {
					if res.uploaded != nil {
						auditLogger.Log(AuditLevelINFO, transactionID, doc.Filename, req.Collection, res.uploaded.DocumentID)
					} else if res.failed != nil {
						outcome := res.failed.Code
						if outcome == "" {
							outcome = types.ErrInternal
						}
						auditLogger.Log(AuditLevelERROR, transactionID, doc.Filename, req.Collection, outcome)
					}
				}
			}(i, doc)
		}
		wg.Wait()

		// 7. Aggregate results.
		uploaded := make([]types.UploadResult, 0, len(req.Documents))
		failed := make([]types.UploadResult, 0)
		for _, res := range results {
			if res.uploaded != nil {
				uploaded = append(uploaded, *res.uploaded)
			} else if res.failed != nil {
				failed = append(failed, *res.failed)
			}
		}

		status := http.StatusOK
		if len(failed) > 0 {
			status = http.StatusMultiStatus
		}

		WriteJSON(w, status, types.UploadResponse{
			Success:    true,
			Collection: req.Collection,
			Uploaded:   uploaded,
			Failed:     failed,
		})
	}
}

// processDocument parses and writes a single document to Firestore.
func processDocument(
	ctx context.Context,
	doc types.Document,
	collection string,
	parse parserFunc,
	fw firestoreWriter,
	maxFileSize int64,
) docResult {
	// Validate content_type.
	if doc.ContentType != "json" && doc.ContentType != "text" {
		return docResult{failed: &types.UploadResult{
			Filename: doc.Filename,
			Status:   "failed",
			Code:     types.ErrInvalidContentType,
			Error:    "content_type must be 'json' or 'text'",
		}}
	}

	// Extract raw content bytes from the decoded JSON value.
	contentBytes, err := extractContentBytes(doc.Content)
	if err != nil {
		return docResult{failed: &types.UploadResult{
			Filename: doc.Filename,
			Status:   "failed",
			Code:     types.ErrParseFailed,
			Error:    "cannot read document content: " + err.Error(),
		}}
	}

	// Parse content using filename extension.
	result := parse(doc.Filename, contentBytes, maxFileSize)
	if result.Document == nil {
		return docResult{failed: &types.UploadResult{
			Filename: doc.Filename,
			Status:   "failed",
			Code:     result.Code,
			Error:    result.Error,
		}}
	}

	// Write to Firestore.
	docID, err := fw.WriteDocument(ctx, collection, result.Document.Content)
	if err != nil {
		code := types.ErrInternal
		var we *firestorepkg.WriteError
		if errors.As(err, &we) {
			code = we.Code
		}
		return docResult{failed: &types.UploadResult{
			Filename: doc.Filename,
			Status:   "failed",
			Code:     code,
			Error:    err.Error(),
		}}
	}

	return docResult{uploaded: &types.UploadResult{
		Filename:   doc.Filename,
		DocumentID: docID,
		Status:     "created",
	}}
}

// extractContentBytes converts the content value decoded from JSON into raw bytes
// suitable for parser.Parse. Strings are used directly; other values are marshalled.
func extractContentBytes(content any) ([]byte, error) {
	switch v := content.(type) {
	case string:
		return []byte(v), nil
	case []byte:
		return v, nil
	case nil:
		return []byte{}, nil
	default:
		return json.Marshal(v)
	}
}
