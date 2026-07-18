package types

import "fmt"

type Document struct {
	Filename    string      `json:"filename"`
	Content     interface{} `json:"content"`
	ContentType string      `json:"content_type"`
}

type UploadRequest struct {
	Collection string     `json:"collection"`
	Documents  []Document `json:"documents"`
}

type UploadResult struct {
	Filename   string `json:"filename"`
	DocumentID string `json:"document_id,omitempty"`
	Status     string `json:"status"`
	Code       string `json:"code,omitempty"`
	Error      string `json:"error,omitempty"`
}

type UploadResponse struct {
	Success    bool           `json:"success"`
	Collection string         `json:"collection"`
	Uploaded   []UploadResult `json:"uploaded"`
	Failed     []UploadResult `json:"failed"`
}

type HealthResponse struct {
	Status        string `json:"status"`
	Version       string `json:"version"`
	ProjectID     string `json:"project_id"`
	AuthMode      string `json:"auth_mode"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type AuthRequest struct {
	IDToken   string `json:"id_token"`
	ProjectID string `json:"project_id"`
}

type AuthResponse struct {
	Success      bool   `json:"success"`
	SessionToken string `json:"session_token"`
	ExpiresIn    int    `json:"expires_in"`
	UID          string `json:"uid"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Error   string `json:"error"`
}

type ShutdownResponse struct {
	Success bool   `json:"success"`
	Status  string `json:"status"`
}

const Version = "0.1.0"

const (
	ErrCollectionRequired        = "COLLECTION_REQUIRED"
	ErrCollectionProjectMismatch = "COLLECTION_PROJECT_MISMATCH"
	ErrInvalidContentType        = "INVALID_CONTENT_TYPE"
	ErrParseFailed               = "PARSE_001"
	ErrEmptyJSON                 = "EMPTY_JSON"
	ErrPayloadTooLarge           = "PAYLOAD_TOO_LARGE"
	ErrAuthInvalidToken          = "AUTH_001"
	ErrAuthSessionInvalid        = "AUTH_002"
	ErrFirestoreQuota            = "FIRESTORE_QUOTA"
	ErrInternal                  = "INTERNAL"
)

func (r *UploadRequest) Validate() error {
	if r.Collection == "" {
		return fmt.Errorf("collection path is required")
	}
	if len(r.Documents) == 0 {
		return fmt.Errorf("at least one document is required")
	}
	for _, doc := range r.Documents {
		if doc.Filename == "" {
			return fmt.Errorf("document filename is required")
		}
		if doc.ContentType != "json" && doc.ContentType != "text" {
			return fmt.Errorf("content_type must be 'json' or 'text'")
		}
	}
	return nil
}
