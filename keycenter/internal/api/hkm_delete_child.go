package api

import (
	"log"
	"net/http"
)

// handleDeleteChild removes a child node by node_id
func (s *Server) handleDeleteChild(w http.ResponseWriter, r *http.Request) {
	nodeID := r.PathValue("node_id")
	if nodeID == "" {
		s.respondError(w, http.StatusBadRequest, "node_id is required")
		return
	}
	if err := s.db.DeleteChild(nodeID); err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	log.Printf("Deleted child node: %s", nodeID)
	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"deleted": nodeID,
	})
}
