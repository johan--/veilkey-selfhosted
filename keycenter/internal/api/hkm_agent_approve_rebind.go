package api

import (
	"net/http"
	"veilkey-keycenter/internal/db"
)

func (s *Server) handleAgentApproveRebind(w http.ResponseWriter, r *http.Request) {
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

	updated, err := s.db.ApproveAgentRebind(agent.NodeID)
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to approve agent rebind: "+err.Error())
		return
	}
	s.saveAuditEvent(
		"vault",
		updated.NodeID,
		"approve_rebind",
		"operator",
		actorIDForRequest(r),
		"",
		"agent_approve_rebind",
		map[string]any{
			"vault_runtime_hash": agent.AgentHash,
			"key_version":        agent.KeyVersion,
			"rebind_required":    agent.RebindRequired,
		},
		map[string]any{
			"vault_runtime_hash": updated.AgentHash,
			"key_version":        updated.KeyVersion,
			"rebind_required":    updated.RebindRequired,
		},
	)

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"status":        "approved",
		"vault_runtime_hash": updated.AgentHash,
		"vault_id":      formatVaultID(updated.VaultName, updated.VaultHash),
		"managed_paths": db.DecodeManagedPaths(updated.ManagedPaths),
		"key_version":   updated.KeyVersion,
	})
}
