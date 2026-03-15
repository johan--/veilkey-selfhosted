package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"veilkey-keycenter/internal/crypto"
)

func TestHKM_Heartbeat_DirectMatch(t *testing.T) {
	srv, handler := setupHKMServer(t)

	childNodeID := crypto.GenerateUUID()
	w := postJSON(handler, "/api/register", map[string]string{
		"node_id": childNodeID, "label": "heartbeat-child", "url": "http://198.51.100.12:10180",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("register: %d", w.Code)
	}

	w = postJSON(handler, "/api/heartbeat", map[string]string{"node_id": childNodeID, "url": "http://198.51.100.12:10180"})
	if w.Code != http.StatusOK {
		t.Fatalf("heartbeat: expected 200, got %d", w.Code)
	}
	var resp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "ok" {
		t.Errorf("status = %v, want ok", resp["status"])
	}

	child, err := srv.db.GetChild(childNodeID)
	if err != nil {
		t.Fatalf("GetChild: %v", err)
	}
	if child.URL != "http://198.51.100.12:10180" {
		t.Errorf("URL = %q", child.URL)
	}
}
