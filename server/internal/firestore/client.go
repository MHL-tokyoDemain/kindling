package firestore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"

	firebase "firebase.google.com/go/v4"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/kindling/kindling/pkg/types"
)

var ErrProjectIDRequired = errors.New("project ID is required; provide --project-id or ensure credentials file has project_id")

type Config struct {
	CredentialsFile string
	ProjectID       string
}

type Client struct {
	firestoreClient *firestore.Client
	projectID       string
}

func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	credsFile, err := resolveCredentials(cfg.CredentialsFile)
	if err != nil {
		return nil, fmt.Errorf("authentication failed: %w", err)
	}

	projectID, err := resolveProjectID(cfg.ProjectID, credsFile)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve project ID: %w", err)
	}

	opt := option.WithCredentialsFile(credsFile)
	app, err := firebase.NewApp(ctx, &firebase.Config{ProjectID: projectID}, opt)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Firebase app: %w", err)
	}

	fClient, err := app.Firestore(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create Firestore client: %w", err)
	}

	return &Client{
		firestoreClient: fClient,
		projectID:       projectID,
	}, nil
}

func resolveCredentials(cfgPath string) (string, error) {
	paths := []string{}
	if cfgPath != "" {
		paths = append(paths, cfgPath)
	}
	if envPath := os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"); envPath != "" {
		paths = append(paths, envPath)
	}
	paths = append(paths, "./serviceAccountKey.json")

	for _, p := range paths {
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}

	return "", errors.New("no service account credentials found: checked --creds, GOOGLE_APPLICATION_CREDENTIALS, and ./serviceAccountKey.json")
}

func resolveProjectID(override string, credsFile string) (string, error) {
	if override != "" {
		return override, nil
	}

	data, err := os.ReadFile(credsFile)
	if err != nil {
		return "", fmt.Errorf("cannot read credentials file: %w", err)
	}

	var sa struct {
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(data, &sa); err != nil {
		return "", fmt.Errorf("invalid credentials JSON: %w", err)
	}

	if sa.ProjectID != "" {
		return sa.ProjectID, nil
	}

	return "", ErrProjectIDRequired
}

func (c *Client) Close() error {
	return c.firestoreClient.Close()
}

func (c *Client) ProjectID() string {
	return c.projectID
}

func (c *Client) WriteDocument(ctx context.Context, collectionPath string, data any) (string, error) {
	docRef, _, err := c.firestoreClient.Collection(collectionPath).Add(ctx, data)
	if err != nil {
		return "", mapFirestoreError(err)
	}
	return docRef.ID, nil
}

type WriteError struct {
	Code    string
	Message string
	Err     error
}

func (e *WriteError) Error() string {
	return e.Message
}

func (e *WriteError) Unwrap() error {
	return e.Err
}

func mapFirestoreError(err error) error {
	st, ok := status.FromError(err)
	if !ok {
		return &WriteError{Code: types.ErrInternal, Message: "unexpected Firestore error", Err: err}
	}

	switch st.Code() {
	case codes.PermissionDenied:
		return &WriteError{Code: types.ErrAuthSessionInvalid, Message: "permission denied", Err: err}
	case codes.ResourceExhausted:
		return &WriteError{Code: types.ErrFirestoreQuota, Message: "Firestore quota exceeded", Err: err}
	default:
		return &WriteError{Code: types.ErrInternal, Message: st.Message(), Err: err}
	}
}

type FirestoreWriter interface {
	WriteDocument(ctx context.Context, collectionPath string, data any) (string, error)
	Close() error
	ProjectID() string
}

var _ FirestoreWriter = (*Client)(nil)
