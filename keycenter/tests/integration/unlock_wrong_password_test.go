package integration_test

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestIntegration_UnlockWrongPassword(t *testing.T) {
	const correct = "the-right-password"
	_, handler, _ := setupServerWithPassword(t, correct)

	w := postJSON(handler, "/api/unlock", map[string]string{"password": "wrong-password-1"})
	if w.Code != http.StatusUnauthorized {
		t.Errorf("wrong password: expected 401, got %d", w.Code)
	}

	var healthResp map[string]string
	json.Unmarshal(getJSON(handler, "/health").Body.Bytes(), &healthResp)
	if healthResp["status"] != "locked" {
		t.Errorf("after wrong: status = %q, want locked", healthResp["status"])
	}

	w = postJSON(handler, "/api/unlock", map[string]string{"password": correct})
	if w.Code != http.StatusOK {
		t.Errorf("correct password: expected 200, got %d", w.Code)
	}
}
