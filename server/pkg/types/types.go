package types

type ServerConfig struct {
	Port        int
	CredsPath   string
	ProjectID   string
	MaxFileSize int64
	Concurrency int
}

type UploadDocument struct {
	Filename    string `json:"filename"`
	Content     string `json:"content"`
	ContentType string `json:"content_type"`
}

type UploadRequest struct {
	Collection string           `json:"collection"`
	Documents  []UploadDocument `json:"documents"`
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
	ProjectID     string `json:"project_id,omitempty"`
	AuthMode      string `json:"auth_mode"`
	UptimeSeconds int64  `json:"uptime_seconds"`
}

type ErrorResponse struct {
	Success bool   `json:"success"`
	Code    string `json:"code"`
	Error   string `json:"error"`
}
