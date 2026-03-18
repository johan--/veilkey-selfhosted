package api

import (
	"encoding/json"
	"net/http"
)

func (s *Server) handleEncrypt(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Plaintext string `json:"plaintext"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		s.respondError(w, http.StatusBadRequest, "invalid request body")
		return
	}
	if req.Plaintext == "" {
		s.respondError(w, http.StatusBadRequest, "plaintext is required")
		return
	}
	s.respondError(w, http.StatusForbidden, vaultcenterOnlyDecryptMessage)
}
