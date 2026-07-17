package server

import (
	"encoding/json"
	"net/http"
)

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	resp := map[string]string{"status": "ok"}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}
