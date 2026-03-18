package api

import (
	"testing"
	"time"

	"veilkey-vaultcenter/internal/db"
)

func TestTrackedRefSyncMovesVersionFromPreviousRef(t *testing.T) {
	srv, handler := setupHKMServer(t)
	first := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"node_id":       "node-ref-sync-1",
		"label":         "ref-sync-agent",
		"vault_hash":    "vh01",
		"vault_name":    "ref-sync-vault",
		"key_version":   7,
		"ip":            "10.0.0.10",
		"port":          10180,
		"secrets_count": 1,
		"configs_count": 0,
	})
	if first.Code != 200 {
		t.Fatalf("agent heartbeat: expected 200, got %d: %s", first.Code, first.Body.String())
	}
	var firstResp struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
		AgentHash        string `json:"agent_hash"`
	}
	decodeJSON(t, first, &firstResp)
	if firstResp.AgentHash != firstResp.VaultRuntimeHash {
		t.Fatalf("runtime hash aliases = %q / %q", firstResp.VaultRuntimeHash, firstResp.AgentHash)
	}

	if err := srv.upsertTrackedRef("VK:TEMP:deadbeef", 7, "temp", firstResp.VaultRuntimeHash); err != nil {
		t.Fatalf("seed temp ref: %v", err)
	}

	w := postJSON(handler, "/api/tracked-refs/sync", map[string]any{
		"vault_runtime_hash": firstResp.VaultRuntimeHash,
		"ref":                "VK:LOCAL:deadbeef",
		"previous_ref":       "VK:TEMP:deadbeef",
		"status":             "active",
	})
	if w.Code != 200 {
		t.Fatalf("sync tracked ref: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	if _, err := srv.db.GetRef("VK:TEMP:deadbeef"); err == nil {
		t.Fatal("previous ref should be removed")
	}
	ref, err := srv.db.GetRef("VK:LOCAL:deadbeef")
	if err != nil {
		t.Fatalf("new ref missing: %v", err)
	}
	if ref.Version != 7 {
		t.Fatalf("version = %d, want 7", ref.Version)
	}
	if ref.Status != "active" {
		t.Fatalf("status = %q, want active", ref.Status)
	}
}

func TestTrackedRefSyncUpdatesStatusInPlace(t *testing.T) {
	srv, handler := setupHKMServer(t)
	first := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"node_id":       "node-ref-sync-2",
		"label":         "config-sync-agent",
		"vault_hash":    "vh02",
		"vault_name":    "config-sync-vault",
		"key_version":   3,
		"ip":            "10.0.0.20",
		"port":          10180,
		"secrets_count": 0,
		"configs_count": 1,
	})
	if first.Code != 200 {
		t.Fatalf("agent heartbeat: expected 200, got %d: %s", first.Code, first.Body.String())
	}
	var firstResp struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
		AgentHash        string `json:"agent_hash"`
	}
	decodeJSON(t, first, &firstResp)
	if firstResp.AgentHash != firstResp.VaultRuntimeHash {
		t.Fatalf("runtime hash aliases = %q / %q", firstResp.VaultRuntimeHash, firstResp.AgentHash)
	}

	if err := srv.upsertTrackedRef("VE:EXTERNAL:APP_URL", 3, "active", firstResp.VaultRuntimeHash); err != nil {
		t.Fatalf("seed external ref: %v", err)
	}

	w := postJSON(handler, "/api/tracked-refs/sync", map[string]any{
		"vault_runtime_hash": firstResp.VaultRuntimeHash,
		"ref":                "VE:EXTERNAL:APP_URL",
		"status":             "revoke",
	})
	if w.Code != 200 {
		t.Fatalf("sync tracked ref: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	ref, err := srv.db.GetRef("VE:EXTERNAL:APP_URL")
	if err != nil {
		t.Fatalf("updated ref missing: %v", err)
	}
	if ref.Version != 3 {
		t.Fatalf("version = %d, want 3", ref.Version)
	}
	if ref.Status != "revoke" {
		t.Fatalf("status = %q, want revoke", ref.Status)
	}
}

func TestTrackedRefSyncCarriesCatalogIdentityFromPreviousRef(t *testing.T) {
	srv, handler := setupHKMServer(t)
	first := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"node_id":       "node-ref-sync-name",
		"label":         "name-owner",
		"vault_hash":    "vh-name",
		"vault_name":    "vault-name",
		"key_version":   5,
		"ip":            "10.0.0.25",
		"port":          10180,
		"secrets_count": 1,
		"configs_count": 0,
	})
	if first.Code != 200 {
		t.Fatalf("agent heartbeat: expected 200, got %d: %s", first.Code, first.Body.String())
	}
	var firstResp struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
	}
	decodeJSON(t, first, &firstResp)

	if err := srv.upsertTrackedRef("VK:TEMP:afd18aa2", 5, "temp", firstResp.VaultRuntimeHash); err != nil {
		t.Fatalf("seed temp ref: %v", err)
	}
	revealedAt := time.Now().UTC().Add(-time.Minute)
	rotatedAt := time.Now().UTC().Add(-2 * time.Minute)
	if err := srv.db.SaveSecretCatalog(&db.SecretCatalog{
		SecretCanonicalID: "vh-name:VK:TEMP:afd18aa2",
		SecretName:        "GEMINI_API_KEY",
		DisplayName:       "Gemini API Key",
		Description:       "Shared external Gemini key",
		TagsJSON:          `["gemini","external"]`,
		Class:             "key",
		Scope:             "TEMP",
		Status:            "temp",
		VaultNodeUUID:     "node-ref-sync-name",
		VaultRuntimeHash:  firstResp.VaultRuntimeHash,
		VaultHash:         "vh-name",
		RefCanonical:      "VK:TEMP:afd18aa2",
		FieldsPresentJSON: `["LOGIN_ID"]`,
		LastRotatedAt:     &rotatedAt,
		LastRevealedAt:    &revealedAt,
	}); err != nil {
		t.Fatalf("SaveSecretCatalog failed: %v", err)
	}

	w := postJSON(handler, "/api/tracked-refs/sync", map[string]any{
		"vault_runtime_hash": firstResp.VaultRuntimeHash,
		"ref":                "VK:EXTERNAL:afd18aa2",
		"previous_ref":       "VK:TEMP:afd18aa2",
		"status":             "active",
	})
	if w.Code != 200 {
		t.Fatalf("sync tracked ref: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	got, err := srv.db.GetSecretCatalogByRef("VK:EXTERNAL:afd18aa2")
	if err != nil {
		t.Fatalf("GetSecretCatalogByRef failed: %v", err)
	}
	if got.SecretName != "GEMINI_API_KEY" {
		t.Fatalf("secret_name = %q, want GEMINI_API_KEY", got.SecretName)
	}
	if got.DisplayName != "Gemini API Key" {
		t.Fatalf("display_name = %q", got.DisplayName)
	}
	if got.Description != "Shared external Gemini key" {
		t.Fatalf("description = %q", got.Description)
	}
	if got.TagsJSON != `["gemini","external"]` {
		t.Fatalf("tags_json = %q", got.TagsJSON)
	}
	if got.Scope != "EXTERNAL" || got.Status != "active" {
		t.Fatalf("scope/status = %s/%s", got.Scope, got.Status)
	}
}

func TestTrackedRefSyncRejectsDifferentAgentOwner(t *testing.T) {
	srv, handler := setupHKMServer(t)

	first := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"node_id":     "node-ref-sync-3",
		"label":       "owner-a",
		"vault_hash":  "vh03",
		"vault_name":  "vault-a",
		"key_version": 2,
		"ip":          "10.0.0.30",
		"port":        10180,
	})
	second := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"node_id":     "node-ref-sync-4",
		"label":       "owner-b",
		"vault_hash":  "vh04",
		"vault_name":  "vault-b",
		"key_version": 2,
		"ip":          "10.0.0.40",
		"port":        10180,
	})
	var firstResp, secondResp struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
		AgentHash        string `json:"agent_hash"`
	}
	decodeJSON(t, first, &firstResp)
	decodeJSON(t, second, &secondResp)
	if firstResp.AgentHash != firstResp.VaultRuntimeHash || secondResp.AgentHash != secondResp.VaultRuntimeHash {
		t.Fatalf("runtime hash aliases mismatch: first=%q/%q second=%q/%q", firstResp.VaultRuntimeHash, firstResp.AgentHash, secondResp.VaultRuntimeHash, secondResp.AgentHash)
	}

	if err := srv.upsertTrackedRef("VK:TEMP:deadbeef", 2, "temp", firstResp.VaultRuntimeHash); err != nil {
		t.Fatalf("seed temp ref: %v", err)
	}

	w := postJSON(handler, "/api/tracked-refs/sync", map[string]any{
		"vault_runtime_hash": secondResp.VaultRuntimeHash,
		"ref":                "VK:LOCAL:deadbeef",
		"previous_ref":       "VK:TEMP:deadbeef",
		"status":             "active",
	})
	if w.Code != 500 {
		t.Fatalf("expected 500, got %d: %s", w.Code, w.Body.String())
	}
}

func TestTrackedRefSyncAcceptsNodeIDTarget(t *testing.T) {
	srv, handler := setupHKMServer(t)
	first := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"node_id":       "node-ref-sync-node-id",
		"label":         "node-id-owner",
		"vault_hash":    "vh05",
		"vault_name":    "vault-node-id",
		"key_version":   9,
		"ip":            "10.0.0.50",
		"port":          10180,
		"secrets_count": 1,
		"configs_count": 0,
	})
	if first.Code != 200 {
		t.Fatalf("agent heartbeat: expected 200, got %d: %s", first.Code, first.Body.String())
	}
	var firstResp struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
	}
	decodeJSON(t, first, &firstResp)
	if err := srv.upsertTrackedRef("VK:TEMP:deadbeef", 9, "temp", firstResp.VaultRuntimeHash); err != nil {
		t.Fatalf("seed temp ref: %v", err)
	}

	w := postJSON(handler, "/api/tracked-refs/sync", map[string]any{
		"node_id":      "node-ref-sync-node-id",
		"ref":          "VK:LOCAL:deadbeef",
		"previous_ref": "VK:TEMP:deadbeef",
		"status":       "active",
	})
	if w.Code != 200 {
		t.Fatalf("sync tracked ref by node_id: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if _, err := srv.db.GetRef("VK:TEMP:deadbeef"); err == nil {
		t.Fatal("previous ref should be removed")
	}
	ref, err := srv.db.GetRef("VK:LOCAL:deadbeef")
	if err != nil {
		t.Fatalf("new ref missing: %v", err)
	}
	if ref.Version != 9 {
		t.Fatalf("version = %d, want 9", ref.Version)
	}
}

func TestTrackedRefSyncWritesAuditEvents(t *testing.T) {
	srv, handler := setupHKMServer(t)
	first := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"node_id":     "node-ref-sync-audit",
		"label":       "audit-owner",
		"vault_hash":  "vh-audit",
		"vault_name":  "vault-audit",
		"key_version": 4,
		"ip":          "10.0.0.60",
		"port":        10180,
	})
	if first.Code != 200 {
		t.Fatalf("agent heartbeat: expected 200, got %d: %s", first.Code, first.Body.String())
	}
	var firstResp struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
	}
	decodeJSON(t, first, &firstResp)

	if err := srv.upsertTrackedRef("VK:TEMP:deadbeef", 4, "temp", firstResp.VaultRuntimeHash); err != nil {
		t.Fatalf("seed temp ref: %v", err)
	}

	w := postJSON(handler, "/api/tracked-refs/sync", map[string]any{
		"vault_runtime_hash": firstResp.VaultRuntimeHash,
		"ref":                "VK:LOCAL:deadbeef",
		"previous_ref":       "VK:TEMP:deadbeef",
		"status":             "active",
	})
	if w.Code != 200 {
		t.Fatalf("sync tracked ref: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	currentRows, err := srv.db.ListAuditEvents("tracked_ref", "VK:LOCAL:deadbeef")
	if err != nil {
		t.Fatalf("ListAuditEvents(current): %v", err)
	}
	if len(currentRows) == 0 || currentRows[0].Action != "sync" {
		t.Fatalf("expected sync audit row, got %+v", currentRows)
	}

	previousRows, err := srv.db.ListAuditEvents("tracked_ref", "VK:TEMP:deadbeef")
	if err != nil {
		t.Fatalf("ListAuditEvents(previous): %v", err)
	}
	if len(previousRows) == 0 || previousRows[0].Action != "delete" {
		t.Fatalf("expected delete audit row, got %+v", previousRows)
	}
}
