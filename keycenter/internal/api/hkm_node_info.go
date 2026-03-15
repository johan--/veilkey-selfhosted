package api

import (
	"net/http"
)

// handleNodeInfo returns this node's identity and stats
func (s *Server) handleNodeInfo(w http.ResponseWriter, r *http.Request) {
	resp, err := s.hkmRuntimeInfo()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "node info not available")
		return
	}
	s.respondJSON(w, http.StatusOK, resp)
}
