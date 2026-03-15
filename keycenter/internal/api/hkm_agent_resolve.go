package api

import (
	"net/http"
	"veilkey-keycenter/internal/crypto"
)

func (s *Server) handleAgentResolve(w http.ResponseWriter, r *http.Request) {
	token := r.PathValue("token")
	if token == "" {
		s.respondError(w, http.StatusBadRequest, "token is required")
		return
	}

	if len(token) <= 8 {
		s.respondError(w, http.StatusBadRequest, "invalid token format: too short")
		return
	}

	agentHash := token[:8]
	secretRef := token[8:]

	agent, err := s.db.GetAgentByHash(agentHash)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "agent not found for hash: "+agentHash)
		return
	}
	if err := validateAgentAvailability(agent); err != nil {
		s.respondAgentLookupError(w, err)
		return
	}

	agentDEK, err := s.decryptAgentDEK(agent.DEK, agent.DEKNonce)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to decrypt agent DEK")
		return
	}

	ai := agentToInfo(agent)
	cipherSecret, err := s.fetchAgentCiphertext(ai.URL(), secretRef)
	if err != nil {
		s.respondError(w, http.StatusNotFound, "failed to fetch secret from agent: "+err.Error())
		return
	}

	plaintext, err := crypto.Decrypt(agentDEK, cipherSecret.Ciphertext, cipherSecret.Nonce)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "decryption failed")
		return
	}

	resp := map[string]interface{}{
		"ref":   secretRef,
		"vault": agent.Label,
		"name":  cipherSecret.Name,
		"value": string(plaintext),
	}
	setRuntimeHashAliases(resp, agentHash)
	s.respondJSON(w, http.StatusOK, resp)
}
