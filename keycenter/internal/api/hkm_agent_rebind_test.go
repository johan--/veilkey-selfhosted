package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"
)

func TestAgentSecretEndpointsRejectRebindRequiredAgent(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "rebind-secret-agent", map[string]string{}, map[string]string{
		"API_KEY": "secret-value",
	})

	agent, err := srv.db.GetAgentByHash(agentHash)
	if err != nil {
		t.Fatalf("GetAgentByHash: %v", err)
	}
	if _, err := srv.db.AdvanceAgentRebind(agent.NodeID, "key_version_mismatch", time.Now().UTC()); err != nil {
		t.Fatalf("AdvanceAgentRebind: %v", err)
	}

	get := getJSON(handler, "/api/agents/"+agentHash+"/secrets/API_KEY")
	if get.Code != http.StatusConflict {
		t.Fatalf("get secret expected 409, got %d: %s", get.Code, get.Body.String())
	}

	save := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "NEW_KEY",
		"value": "new-secret",
	})
	if save.Code != http.StatusConflict {
		t.Fatalf("save secret expected 409, got %d: %s", save.Code, save.Body.String())
	}
}

func TestAgentResolveRejectsBlockedAgent(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "blocked-resolve-agent", map[string]string{}, map[string]string{
		"API_KEY": "secret-value",
	})

	agent, err := srv.db.GetAgentByHash(agentHash)
	if err != nil {
		t.Fatalf("GetAgentByHash: %v", err)
	}
	for i := 0; i < 4; i++ {
		if _, err := srv.db.AdvanceAgentRebind(agent.NodeID, "key_version_mismatch", time.Now().UTC()); err != nil {
			t.Fatalf("AdvanceAgentRebind %d: %v", i+1, err)
		}
	}

	resolve := getJSON(handler, "/api/resolve-agent/"+agentHash+"deadbeef")
	if resolve.Code != http.StatusLocked {
		t.Fatalf("resolve expected 423, got %d: %s", resolve.Code, resolve.Body.String())
	}
}

func TestAgentApproveRebindRotatesHashAndVersion(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "approve-rebind-agent", map[string]string{}, map[string]string{
		"API_KEY": "secret-value",
	})

	agent, err := srv.db.GetAgentByHash(agentHash)
	if err != nil {
		t.Fatalf("GetAgentByHash: %v", err)
	}
	if _, err := srv.db.AdvanceAgentRebind(agent.NodeID, "key_version_mismatch", time.Now().UTC()); err != nil {
		t.Fatalf("AdvanceAgentRebind: %v", err)
	}

	w := postJSON(handler, "/api/agents/"+agentHash+"/approve-rebind", map[string]string{})
	if w.Code != http.StatusOK {
		t.Fatalf("approve rebind expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Status           string `json:"status"`
		VaultRuntimeHash string `json:"vault_runtime_hash"`
		KeyVersion       int    `json:"key_version"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal approve response: %v", err)
	}
	if resp.Status != "approved" {
		t.Fatalf("status = %q", resp.Status)
	}
	if resp.VaultRuntimeHash == "" || resp.VaultRuntimeHash == agentHash {
		t.Fatalf("vault_runtime_hash = %q, old=%q", resp.VaultRuntimeHash, agentHash)
	}
	if resp.KeyVersion != 2 {
		t.Fatalf("key_version = %d", resp.KeyVersion)
	}

	updated, err := srv.db.GetAgentByNodeID(agent.NodeID)
	if err != nil {
		t.Fatalf("GetAgentByNodeID: %v", err)
	}
	if updated.RebindRequired {
		t.Fatal("rebind_required should be cleared")
	}
	if updated.BlockedAt != nil {
		t.Fatal("blocked_at should be cleared")
	}
	if updated.KeyVersion != 2 {
		t.Fatalf("stored key_version = %d", updated.KeyVersion)
	}
	auditRows, err := srv.db.ListAuditEvents("vault", agent.NodeID)
	if err != nil {
		t.Fatalf("ListAuditEvents: %v", err)
	}
	if len(auditRows) == 0 || auditRows[0].Action != "approve_rebind" {
		t.Fatalf("expected approve_rebind audit, got %+v", auditRows)
	}
	if auditRows[0].ActorID == "" {
		t.Fatalf("expected operator actor id in audit row, got %+v", auditRows[0])
	}
}

func TestAgentRebindPlanReturnsNextKeyVersion(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "plan-rebind-agent", map[string]string{}, map[string]string{
		"API_KEY": "secret-value",
	})

	agent, err := srv.db.GetAgentByHash(agentHash)
	if err != nil {
		t.Fatalf("GetAgentByHash: %v", err)
	}
	if _, err := srv.db.AdvanceAgentRebind(agent.NodeID, "key_version_mismatch", time.Now().UTC()); err != nil {
		t.Fatalf("AdvanceAgentRebind: %v", err)
	}

	w := getJSON(handler, "/api/agents/"+agentHash+"/rebind-plan")
	if w.Code != http.StatusOK {
		t.Fatalf("rebind plan expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Status            string `json:"status"`
		CurrentKeyVersion int    `json:"current_key_version"`
		NextKeyVersion    int    `json:"next_key_version"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal plan response: %v", err)
	}
	if resp.Status != "plan" {
		t.Fatalf("status = %q", resp.Status)
	}
	if resp.CurrentKeyVersion != 1 || resp.NextKeyVersion != 2 {
		t.Fatalf("versions = current:%d next:%d", resp.CurrentKeyVersion, resp.NextKeyVersion)
	}
}
