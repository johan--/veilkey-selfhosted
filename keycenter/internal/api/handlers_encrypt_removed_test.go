package api

import (
	"net/http"
	"testing"
)

func TestEncryptEndpointRemoved(t *testing.T) {
	_, handler := setupTestServer(t)

	w := postJSON(handler, "/api/encrypt", map[string]string{"plaintext": "test"})
	if w.Code != http.StatusNotFound {
		t.Errorf("POST /api/encrypt should return 404, got %d", w.Code)
	}
}
