package server

import (
	"encoding/json"
	"net/http"
)

type statusResponse struct {
	Status  string `json:"status"`
	Version string `json:"version"`
}

func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(statusResponse{ //nolint:errcheck
		Status:  "ok",
		Version: version,
	})
}
