package api

import (
	"net/http"
	"veilkey-vaultcenter/internal/db"
)

func (s *Server) handleAgentRebindPlan(w http.ResponseWriter, r *http.Request) {
	hashOrLabel := r.PathValue("agent")
	agent, err := s.findAgentRecord(hashOrLabel)
	if err != nil {
		s.respondError(w, http.StatusNotFound, err.Error())
		return
	}
	if !agent.RebindRequired && agent.BlockedAt == nil {
		s.respondError(w, http.StatusBadRequest, "agent does not require rebind")
		return
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":              "plan",
		"vault_runtime_hash":  agent.AgentHash,
		"vault_id":            formatVaultID(agent.VaultName, agent.VaultHash),
		"current_key_version": agent.KeyVersion,
		"next_key_version":    agent.KeyVersion + 1,
		"managed_paths":       db.DecodeManagedPaths(agent.ManagedPaths),
		"rebind_required":     agent.RebindRequired,
		"blocked":             agent.BlockedAt != nil,
	})
}
