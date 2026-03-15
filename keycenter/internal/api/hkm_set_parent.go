package api

import (
	"encoding/json"
	"log"
	"net/http"
)

// handleSetParent sets the parent_url in node_info
func (s *Server) handleSetParent(w http.ResponseWriter, r *http.Request) {
	var req struct {
		ParentURL string `json:"parent_url"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.ParentURL == "" {
		s.respondError(w, http.StatusBadRequest, "parent_url is required")
		return
	}
	_, err := s.db.SetParentURL(req.ParentURL)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, err.Error())
		return
	}
	log.Printf("Parent URL set to %s", req.ParentURL)
	s.respondJSON(w, http.StatusOK, map[string]interface{}{"status": "ok", "parent_url": req.ParentURL})
}
