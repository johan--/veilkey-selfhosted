package api

import (
	"net/http"
)

// handleListRegistry returns key distribution map
func (s *Server) handleListRegistry(w http.ResponseWriter, r *http.Request) {
	entries, err := s.db.ListRegistry()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list registry")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"entries": entries,
		"count":   len(entries),
	})
}
