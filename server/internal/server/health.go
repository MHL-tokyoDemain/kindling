package server

import (
	"net/http"
	"time"

	"github.com/kindling/kindling/pkg/types"
)

func (s *Server) HealthHandler(w http.ResponseWriter, r *http.Request) {
	uptime := int64(time.Since(s.startTime).Seconds())

	resp := types.HealthResponse{
		Status:        "ok",
		Version:       types.Version,
		ProjectID:     s.cfg.ProjectID,
		AuthMode:      "service_account",
		UptimeSeconds: uptime,
	}
	WriteJSON(w, http.StatusOK, resp)
}
