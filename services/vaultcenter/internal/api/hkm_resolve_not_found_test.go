package api

import (
	"net/http"
	"testing"
)

func TestHKM_ResolveNotFound(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/resolve/nonexistent")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
