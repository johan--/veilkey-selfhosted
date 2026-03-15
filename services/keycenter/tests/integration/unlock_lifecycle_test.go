package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestIntegration_UnlockLifecycle(t *testing.T) {
	const password = "correct-horse-battery-staple"
	_, handler, _ := setupServerWithPassword(t, password)

	var healthResp map[string]string
	json.Unmarshal(getJSON(handler, "/health").Body.Bytes(), &healthResp)
	if healthResp["status"] != "locked" {
		t.Errorf("before unlock: status = %q, want locked", healthResp["status"])
	}

	w := getJSON(handler, "/api/status")
	if w.Code != http.StatusServiceUnavailable {
		t.Errorf("status before unlock: expected 503, got %d", w.Code)
	}

	w = postJSON(handler, "/api/unlock", map[string]string{"password": password})
	if w.Code != http.StatusOK {
		t.Fatalf("unlock: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var unlockResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &unlockResp)
	if unlockResp["status"] != "unlocked" {
		t.Errorf("unlock status = %v, want unlocked", unlockResp["status"])
	}

	json.Unmarshal(getJSON(handler, "/health").Body.Bytes(), &healthResp)
	if healthResp["status"] != "ok" {
		t.Errorf("after unlock: status = %q, want ok", healthResp["status"])
	}

	w = postJSON(handler, "/api/unlock", map[string]string{"password": password})
	if w.Code != http.StatusOK {
		t.Fatalf("second unlock: expected 200, got %d", w.Code)
	}
	var secondResp map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &secondResp)
	if secondResp["status"] != "already_unlocked" {
		t.Errorf("second unlock status = %v, want already_unlocked", secondResp["status"])
	}
}
