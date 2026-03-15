package api

import (
	"net/http"
	"testing"
)

func TestDecryptEndpointRemoved(t *testing.T) {
	_, handler := setupTestServer(t)

	w := postJSON(handler, "/api/decrypt", map[string]string{"ciphertext": "VK:TEMP:deadbeef"})
	if w.Code != http.StatusNotFound {
		t.Errorf("POST /api/decrypt should return 404, got %d", w.Code)
	}
}
