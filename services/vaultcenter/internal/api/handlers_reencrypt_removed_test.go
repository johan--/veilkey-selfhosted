package api

import (
	"net/http"
	"testing"
)

func TestReencryptEndpointRemoved(t *testing.T) {
	_, handler := setupTestServer(t)

	w := postJSON(handler, "/api/reencrypt", map[string]string{"ciphertext": "VK:TEMP:deadbeef"})
	if w.Code != http.StatusNotFound {
		t.Errorf("POST /api/reencrypt should return 404, got %d", w.Code)
	}
}
