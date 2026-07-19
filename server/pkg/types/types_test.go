package types

import (
	"testing"
)

func TestValidate(t *testing.T) {
	tests := []struct {
		name    string
		req     *UploadRequest
		wantErr string
	}{
		{
			name:    "empty collection",
			req:     &UploadRequest{},
			wantErr: "collection path is required",
		},
		{
			name:    "no documents",
			req:     &UploadRequest{Collection: "test-col"},
			wantErr: "at least one document is required",
		},
		{
			name: "empty filename",
			req: &UploadRequest{
				Collection: "test-col",
				Documents:  []Document{{Filename: "", ContentType: "json"}},
			},
			wantErr: "document filename is required",
		},
		{
			name: "invalid content type",
			req: &UploadRequest{
				Collection: "test-col",
				Documents:  []Document{{Filename: "f.json", ContentType: "xml"}},
			},
			wantErr: "content_type must be 'json' or 'text'",
		},
		{
			name: "valid json",
			req: &UploadRequest{
				Collection: "test-col",
				Documents:  []Document{{Filename: "f.json", ContentType: "json"}},
			},
			wantErr: "",
		},
		{
			name: "valid text",
			req: &UploadRequest{
				Collection: "test-col",
				Documents:  []Document{{Filename: "f.txt", ContentType: "text"}},
			},
			wantErr: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if tt.wantErr == "" {
				if err != nil {
					t.Errorf("expected no error, got %v", err)
				}
			} else {
				if err == nil {
					t.Errorf("expected error %q, got nil", tt.wantErr)
				} else if err.Error() != tt.wantErr {
					t.Errorf("expected error %q, got %q", tt.wantErr, err.Error())
				}
			}
		})
	}
}
