package api

import (
	"net/http"
	"testing"
)

func TestHKM_Heartbeat_MissingFields(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := postJSON(handler, "/api/heartbeat", map[string]string{"node_id": "test"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing url: expected 400, got %d", w.Code)
	}

	w = postJSON(handler, "/api/heartbeat", map[string]string{"url": "http://example.com"})
	if w.Code != http.StatusBadRequest {
		t.Errorf("missing node_id: expected 400, got %d", w.Code)
	}
}
