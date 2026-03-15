package api

import (
	"net/http"
	"testing"
)

func TestHKM_Heartbeat_EmptyBody(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := postJSON(handler, "/api/heartbeat", map[string]string{})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
