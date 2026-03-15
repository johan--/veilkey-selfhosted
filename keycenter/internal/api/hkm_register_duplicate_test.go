package api

import (
	"net/http"
	"testing"
	"veilkey-keycenter/internal/crypto"
)

func TestHKM_Register_Duplicate(t *testing.T) {
	_, handler := setupHKMServer(t)

	nodeID := crypto.GenerateUUID()
	w := postJSON(handler, "/api/register", map[string]string{"node_id": nodeID, "label": "first"})
	if w.Code != http.StatusOK {
		t.Fatalf("first register: %d", w.Code)
	}

	w = postJSON(handler, "/api/register", map[string]string{"node_id": nodeID, "label": "second"})
	if w.Code != http.StatusConflict {
		t.Errorf("duplicate: expected 409, got %d", w.Code)
	}
}
