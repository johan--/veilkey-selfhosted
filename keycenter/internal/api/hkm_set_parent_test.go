package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHKM_SetParent(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := postJSON(handler, "/api/set-parent", map[string]string{"parent_url": "http://198.51.100.1:10180"})
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	w = getJSON(handler, "/api/node-info")
	var resp struct {
		ParentURL string `json:"parent_url"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.ParentURL != "http://198.51.100.1:10180" {
		t.Errorf("parent_url = %q", resp.ParentURL)
	}
}
