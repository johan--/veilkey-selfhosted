package api

import (
	"log"
	"net/http"
)

func (s *Server) handleAgentUnregisterByNode(w http.ResponseWriter, r *http.Request) {
	nodeID := r.PathValue("node_id")
	if nodeID == "" {
		s.respondError(w, http.StatusBadRequest, "node_id is required")
		return
	}
	if err := s.db.DeleteAgentByNodeID(nodeID); err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	log.Printf("agent: unregistered node=%s", nodeID)
	s.respondJSON(w, http.StatusOK, map[string]any{
		"deleted": nodeID,
		"status":  "unregistered",
	})
}
