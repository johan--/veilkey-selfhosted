package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"veilkey-vaultcenter/internal/crypto"
)

func TestHKM_ResolveFederatedChild(t *testing.T) {
	_, handler := setupHKMServer(t)

	mockChild := newMockChildResolve(t)
	defer mockChild.Close()

	childNodeID := crypto.GenerateUUID()
	w := postJSON(handler, "/api/register", map[string]interface{}{
		"node_id": childNodeID, "label": "mock-child", "url": mockChild.URL,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("register: %d: %s", w.Code, w.Body.String())
	}

	postJSON(handler, "/api/heartbeat", map[string]string{"node_id": childNodeID, "url": mockChild.URL})

	w = getJSON(handler, "/api/resolve/test_ref")
	if w.Code != http.StatusOK {
		t.Fatalf("federated resolve: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Value string `json:"value"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Value != "from-child" {
		t.Errorf("value = %q, want from-child", resp.Value)
	}
}
