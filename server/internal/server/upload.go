package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) UploadHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "not_implemented"}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(resp)
}
