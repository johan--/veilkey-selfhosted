package api

import (
	"net/http"
	"testing"
)

func TestHKM_Rekey_EmptyBody(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := postJSON(handler, "/api/rekey", map[string]interface{}{})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}
