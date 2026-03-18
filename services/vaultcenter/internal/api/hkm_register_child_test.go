package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"veilkey-vaultcenter/internal/crypto"
)

func TestHKM_RegisterChild(t *testing.T) {
	_, handler := setupHKMServer(t)

	childNodeID := crypto.GenerateUUID()
	w := postJSON(handler, "/api/register", map[string]string{"node_id": childNodeID, "label": "test-child"})
	if w.Code != http.StatusOK {
		t.Fatalf("register: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		DEK     []byte `json:"dek"`
		Version int    `json:"version"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if len(resp.DEK) != 32 {
		t.Errorf("DEK length = %d, want 32", len(resp.DEK))
	}
	if resp.Version != 1 {
		t.Errorf("version = %d, want 1", resp.Version)
	}

	w = getJSON(handler, "/api/children")
	if w.Code != http.StatusOK {
		t.Fatalf("children: expected 200, got %d", w.Code)
	}
	var childResp struct {
		Count int `json:"count"`
	}
	json.Unmarshal(w.Body.Bytes(), &childResp)
	if childResp.Count != 1 {
		t.Errorf("count = %d, want 1", childResp.Count)
	}

	w = postJSON(handler, "/api/register", map[string]string{"node_id": childNodeID, "label": "duplicate"})
	if w.Code != http.StatusConflict {
		t.Errorf("duplicate: expected 409, got %d", w.Code)
	}
}
