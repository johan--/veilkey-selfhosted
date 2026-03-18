package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestStatusEndpoint(t *testing.T) {
	_, handler := setupTestServer(t)

	w := getJSON(handler, "/api/status")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Mode             string `json:"mode"`
		NodeID           string `json:"node_id"`
		VaultNodeUUID    string `json:"vault_node_uuid"`
		Version          int    `json:"version"`
		ChildrenCount    int    `json:"children_count"`
		TrackedRefsCount int    `json:"tracked_refs_count"`
		Locked           bool   `json:"locked"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp.Mode != "hkm" {
		t.Errorf("mode = %q, want hkm", resp.Mode)
	}
	if resp.NodeID == "" {
		t.Error("node_id should not be empty")
	}
	if resp.VaultNodeUUID == "" {
		t.Error("vault_node_uuid should not be empty")
	}
	if resp.VaultNodeUUID != resp.NodeID {
		t.Errorf("vault_node_uuid = %q, want %q", resp.VaultNodeUUID, resp.NodeID)
	}
	if resp.Version < 1 {
		t.Errorf("version = %d, expected >= 1", resp.Version)
	}
	if resp.TrackedRefsCount != 0 {
		t.Errorf("tracked_refs_count = %d, want 0", resp.TrackedRefsCount)
	}
	if resp.Locked {
		t.Error("locked should be false")
	}
}
