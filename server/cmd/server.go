package cmd

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"github.com/kindling/kindling/internal/firestore"
	"github.com/kindling/kindling/internal/parser"
	"github.com/kindling/kindling/internal/server"
	"github.com/kindling/kindling/pkg/types"
)

func RunServer(port int, credsPath, projectID string, maxFileSize int64) error {
	ctx := context.Background()

	// Resolve project ID before starting the server.
	resolvedProjectID, err := server.ResolveProjectID(ctx, projectID, credsPath)
	if err != nil {
		return fmt.Errorf("project ID resolution failed: %w", err)
	}

	// Create the Firestore client that the upload handler will use.
	fw, err := firestore.NewClient(ctx, firestore.Config{
		CredentialsFile: credsPath,
		ProjectID:       resolvedProjectID,
	})
	if err != nil {
		return fmt.Errorf("failed to create Firestore client: %w", err)
	}

	cfg := server.Config{
		Port:            port,
		CredentialsFile: credsPath,
		ProjectID:       resolvedProjectID,
		MaxFileSize:     maxFileSize,
	}

	srv, err := server.New(cfg, fw)
	if err != nil {
		return fmt.Errorf("failed to create server: %w", err)
	}

	return srv.Start()
}

func RunUpload(collection string, files []string, credsPath, projectID string, maxFileSize int64, concurrency int) (int, error) {
	if collection == "" {
		return 2, fmt.Errorf("--collection is required")
	}
	if len(files) == 0 {
		return 2, fmt.Errorf("--files is required")
	}

	ctx := context.Background()

	firestoreCfg := firestore.Config{
		CredentialsFile: credsPath,
		ProjectID:       projectID,
	}
	firestoreClient, err := firestore.NewClient(ctx, firestoreCfg)
	if err != nil {
		return 3, fmt.Errorf("failed to create Firestore client: %w", err)
	}
	defer firestoreClient.Close()

	uploaded := make([]types.UploadResult, 0)
	failed := make([]types.UploadResult, 0)

	for _, filePath := range files {
		content, err := os.ReadFile(filePath)
		if err != nil {
			failed = append(failed, types.UploadResult{
				Filename: filePath,
				Status:   "failed",
				Code:     "",
				Error:    fmt.Sprintf("cannot read file: %v", err),
			})
			continue
		}

		result := parser.Parse(filePath, content, maxFileSize)
		if result.Document == nil {
			failed = append(failed, types.UploadResult{
				Filename: filePath,
				Status:   "failed",
				Code:     result.Code,
				Error:    result.Error,
			})
			continue
		}

		docID, err := firestoreClient.WriteDocument(ctx, collection, result.Document)
		if err != nil {
			var writeErr *firestore.WriteError
			code := ""
			if errors.As(err, &writeErr) {
				code = writeErr.Code
			}
			failed = append(failed, types.UploadResult{
				Filename: filePath,
				Status:   "failed",
				Code:     code,
				Error:    err.Error(),
			})
			continue
		}

		uploaded = append(uploaded, types.UploadResult{
			Filename:   filePath,
			DocumentID: docID,
			Status:     "created",
		})
	}

	resp := types.UploadResponse{
		Success:    len(failed) == 0,
		Collection: collection,
		Uploaded:   uploaded,
		Failed:     failed,
	}

	out, err := json.MarshalIndent(resp, "", "  ")
	if err != nil {
		return 1, fmt.Errorf("failed to encode output: %w", err)
	}
	fmt.Println(string(out))

	if len(failed) > 0 {
		return 1, nil
	}
	return 0, nil
}
