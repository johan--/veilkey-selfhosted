package api

import (
	"net/http"
	"testing"
	"veilkey-vaultcenter/internal/crypto"
)

func TestHKM_Heartbeat_VersionChain(t *testing.T) {
	_, handler := setupHKMServer(t)

	nodeID := crypto.GenerateUUID()
	url := "http://198.51.100.13:10180"

	w := postJSON(handler, "/api/register", map[string]string{"node_id": nodeID, "label": "version-child", "url": url})
	if w.Code != http.StatusOK {
		t.Fatalf("register: %d", w.Code)
	}

	w = postJSON(handler, "/api/heartbeat", map[string]interface{}{"node_id": nodeID, "url": url, "version": 1})
	if w.Code != http.StatusOK {
		t.Fatalf("heartbeat v1: expected 200, got %d", w.Code)
	}

	w = postJSON(handler, "/api/heartbeat", map[string]interface{}{"node_id": nodeID, "url": url, "version": 99})
	if w.Code != http.StatusForbidden {
		t.Errorf("wrong version: expected 403, got %d", w.Code)
	}

	w = postJSON(handler, "/api/heartbeat", map[string]interface{}{"node_id": nodeID, "url": url, "version": 1})
	if w.Code != http.StatusNotFound {
		t.Errorf("after disconnect: expected 404, got %d", w.Code)
	}
}
