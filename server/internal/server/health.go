package server

import (
	"encoding/json"
	"net/http"

	"github.com/kindling/kindling/pkg/types"
)

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	resp := types.HealthResponse{
		Status:  "ok",
		Version: "0.1.0",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(resp)
}
