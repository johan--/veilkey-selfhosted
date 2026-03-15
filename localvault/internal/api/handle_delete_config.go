package api

import "net/http"

func (s *Server) handleDeleteConfig(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	if key == "" {
		s.respondError(w, http.StatusBadRequest, "key is required")
		return
	}

	if err := s.db.DeleteConfig(key); err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"deleted": key,
	})
}
