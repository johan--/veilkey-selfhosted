package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHKM_ResolveCascadeHeader(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := postJSONWithHeader(handler, "GET", "/api/resolve/unknown_ref", map[string]string{
		"X-VeilKey-Cascade": "true",
	})
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
	var resp map[string]string
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp["error"] != "ref not found locally: unknown_ref" {
		t.Errorf("error = %q", resp["error"])
	}
}
