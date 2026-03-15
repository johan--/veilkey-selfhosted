package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHKM_StatusMode(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/status")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Mode          string `json:"mode"`
		NodeID        string `json:"node_id"`
		VaultNodeUUID string `json:"vault_node_uuid"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
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
}
