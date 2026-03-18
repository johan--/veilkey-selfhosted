package api

import (
	"net/http"
	"testing"
)

func TestHKM_Rekey_InvalidDEKLength(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := postJSON(handler, "/api/rekey", map[string]interface{}{"dek": make([]byte, 16), "version": 2})
	if w.Code != http.StatusBadRequest {
		t.Errorf("short DEK: expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
