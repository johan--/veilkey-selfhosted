package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHKM_NodeInfo(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/node-info")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		NodeID           string `json:"node_id"`
		VaultNodeUUID    string `json:"vault_node_uuid"`
		Version          int    `json:"version"`
		Mode             string `json:"mode"`
		ChildrenCount    int    `json:"children_count"`
		TrackedRefsCount int    `json:"tracked_refs_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode node-info: %v", err)
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
	if resp.Version != 1 {
		t.Errorf("version = %d, want 1", resp.Version)
	}
	if resp.ChildrenCount != 0 {
		t.Errorf("children_count = %d, want 0", resp.ChildrenCount)
	}
	if resp.TrackedRefsCount != 0 {
		t.Errorf("tracked_refs_count = %d, want 0", resp.TrackedRefsCount)
	}
}
