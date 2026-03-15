package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"veilkey-keycenter/internal/crypto"
)

func TestHKM_DeleteChild(t *testing.T) {
	_, handler := setupHKMServer(t)

	childNodeID := crypto.GenerateUUID()
	w := postJSON(handler, "/api/register", map[string]string{"node_id": childNodeID, "label": "delete-me"})
	if w.Code != http.StatusOK {
		t.Fatalf("register: %d", w.Code)
	}

	w = deleteJSON(handler, "/api/children/"+childNodeID)
	if w.Code != http.StatusOK {
		t.Fatalf("delete: expected 200, got %d", w.Code)
	}
	var resp struct {
		Deleted string `json:"deleted"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Deleted != childNodeID {
		t.Errorf("deleted = %q, want %q", resp.Deleted, childNodeID)
	}

	w = deleteJSON(handler, "/api/children/nonexistent-node")
	if w.Code != http.StatusNotFound {
		t.Errorf("delete missing: expected 404, got %d", w.Code)
	}
}
