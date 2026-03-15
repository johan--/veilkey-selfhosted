package api

import (
	"net/http"
	"testing"
)

func TestHKM_SetParent_Empty(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := postJSON(handler, "/api/set-parent", map[string]string{"parent_url": ""})
	if w.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", w.Code)
	}
}
