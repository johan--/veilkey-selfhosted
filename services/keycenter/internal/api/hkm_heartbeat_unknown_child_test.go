package api

import (
	"net/http"
	"testing"
	"veilkey-keycenter/internal/crypto"
)

func TestHKM_Heartbeat_UnknownChild(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := postJSON(handler, "/api/heartbeat", map[string]string{
		"node_id": crypto.GenerateUUID(), "url": "http://198.51.100.99:10180",
	})
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
