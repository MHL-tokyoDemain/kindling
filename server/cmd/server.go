package cmd

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/kindling/kindling/internal/firestore"
	"github.com/kindling/kindling/internal/parser"
	"github.com/kindling/kindling/internal/server"
	"github.com/kindling/kindling/pkg/types"
)

func RunServer(port int, credsPath, projectID string, maxFileSize int64) error {
	srv := server.New(port)
	// TODO: pass credsPath, projectID, maxFileSize to server config (issue #11)
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
	_ = projectID
	_ = concurrency

	// TODO(#7): Initialize Firestore client with proper project binding (issue #10)
	firestoreClient, err := firestore.NewClient(ctx, credsPath)
	if err != nil {
		return 3, fmt.Errorf("authentication failed: %w", err)
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

		// TODO(#10): Write document to Firestore via firestoreClient and capture document ID
		_ = firestoreClient

		uploaded = append(uploaded, types.UploadResult{
			Filename:   filePath,
			DocumentID: "",
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
