package server

import (
	"io"
	"net/http"

	"github.com/kindling/kindling/pkg/types"
)

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxFileSize)

	body, err := io.ReadAll(r.Body)
	if err != nil {
		writeJSON(w, http.StatusRequestEntityTooLarge, types.ErrorResponse{
			Success: false,
			Code:    types.ErrPayloadTooLarge,
			Error:   "request body exceeds size limit",
		})
		return
	}
	_ = body

	// TODO(#13): Implement real upload handler with parsing, validation, Firestore writes
	writeJSON(w, http.StatusNotImplemented, map[string]string{"status": "not_implemented"})
}
