package server

import (
	"errors"
	"io"
	"net/http"

	"github.com/kindling/kindling/pkg/types"
)

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, s.cfg.MaxFileSize)

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
	_ = body

	// TODO(#13): Implement real upload handler with parsing, validation, Firestore writes
	WriteJSON(w, http.StatusNotImplemented, map[string]string{"status": "not_implemented"})
}
