package api

import "testing"

func TestTrackedRefCleanupPreviewAndApply(t *testing.T) {
	srv, handler := setupHKMServer(t)

	hb := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"vault_node_uuid": "node-cleanup-a",
		"label":           "cleanup-agent-a",
		"vault_hash":      "vh-cleanup-a",
		"vault_name":      "cleanup-a",
		"key_version":     5,
		"ip":              "10.0.1.31",
		"port":            10180,
	})
	var agentA struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
	}
	decodeJSON(t, hb, &agentA)

	hb = postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"vault_node_uuid": "node-cleanup-b",
		"label":           "cleanup-agent-b",
		"vault_hash":      "vh-cleanup-b",
		"vault_name":      "cleanup-b",
		"key_version":     5,
		"ip":              "10.0.1.32",
		"port":            10181,
	})
	var agentB struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
	}
	decodeJSON(t, hb, &agentB)

	for _, seed := range []struct {
		ref    string
		status string
		owner  string
	}{
		{"VK:TEMP:dup001", "temp", agentA.VaultRuntimeHash},
		{"VK:LOCAL:dup001", "active", agentA.VaultRuntimeHash},
		{"VE:LOCAL:APP_URL", "active", ""},
		{"VK:LOCAL:shared001", "active", agentA.VaultRuntimeHash},
		{"VK:TEMP:shared001", "temp", agentB.VaultRuntimeHash},
	} {
		if err := srv.upsertTrackedRef(seed.ref, 5, seed.status, seed.owner); err != nil {
			t.Fatalf("seed %s: %v", seed.ref, err)
		}
	}

	preview := postJSON(handler, "/api/tracked-refs/cleanup", map[string]any{})
	if preview.Code != 200 {
		t.Fatalf("preview expected 200, got %d: %s", preview.Code, preview.Body.String())
	}
	var previewResp struct {
		Status  string `json:"status"`
		Counts  map[string]int
		Actions []struct {
			Reason string   `json:"reason"`
			Delete []string `json:"delete"`
			Keep   []string `json:"keep"`
			Manual bool     `json:"manual"`
		} `json:"actions"`
	}
	decodeJSON(t, preview, &previewResp)
	if previewResp.Status != "preview" {
		t.Fatalf("status = %q, want preview", previewResp.Status)
	}
	if previewResp.Counts["delete_candidates"] != 2 {
		t.Fatalf("delete_candidates = %d, want 2", previewResp.Counts["delete_candidates"])
	}
	if previewResp.Counts["manual_actions"] != 1 {
		t.Fatalf("manual_actions = %d, want 1", previewResp.Counts["manual_actions"])
	}

	applied := postJSON(handler, "/api/tracked-refs/cleanup", map[string]any{"apply": true})
	if applied.Code != 200 {
		t.Fatalf("apply expected 200, got %d: %s", applied.Code, applied.Body.String())
	}
	var appliedResp struct {
		Status string         `json:"status"`
		Counts map[string]int `json:"counts"`
	}
	decodeJSON(t, applied, &appliedResp)
	if appliedResp.Status != "applied" {
		t.Fatalf("status = %q, want applied", appliedResp.Status)
	}
	if appliedResp.Counts["deleted"] != 2 {
		t.Fatalf("deleted = %d, want 2", appliedResp.Counts["deleted"])
	}

	audit := getJSON(handler, "/api/tracked-refs/audit")
	var auditResp struct {
		Stale []trackedRefAuditIssue `json:"stale"`
	}
	decodeJSON(t, audit, &auditResp)
	if len(auditResp.Stale) != 1 {
		t.Fatalf("stale issues = %d, want 1 (%+v)", len(auditResp.Stale), auditResp.Stale)
	}
	if auditResp.Stale[0].Reason != "agent_mismatch" {
		t.Fatalf("remaining stale reason = %q, want agent_mismatch", auditResp.Stale[0].Reason)
	}
}
