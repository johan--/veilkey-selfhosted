package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestLockedServerRejects(t *testing.T) {
	_, handler := setupLockedServer(t)

	w := getJSON(handler, "/api/status")
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status on locked: expected 503, got %d", w.Code)
	}

	w = getJSON(handler, "/health")
	if w.Code != http.StatusOK {
		t.Errorf("health on locked: expected 200, got %d", w.Code)
	}
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["status"] != "locked" {
		t.Errorf("health status = %q, want locked", resp["status"])
	}

	w = getJSON(handler, "/ready")
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("ready on locked: expected 503, got %d", w.Code)
	}
}
