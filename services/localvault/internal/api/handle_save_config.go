package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleSaveConfig(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Key   string  `json:"key"`
		Value *string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Key == "" {
		s.respondError(w, http.StatusBadRequest, "key is required")
		return
	}
	if !isValidResourceName(req.Key) {
		s.respondError(w, http.StatusBadRequest, "key must match [A-Z_][A-Z0-9_]*")
		return
	}
	if req.Value == nil {
		s.respondError(w, http.StatusBadRequest, "value is required (use DELETE to remove a config)")
		return
	}

	if err := s.db.SaveConfig(req.Key, *req.Value); err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to save config: "+err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"key":    req.Key,
		"value":  *req.Value,
		"ref":    ParsedRef{Family: RefFamilyVE, Scope: RefScopeLocal, ID: req.Key}.CanonicalString(),
		"scope":  "LOCAL",
		"status": "active",
		"action": "saved",
	})
}
