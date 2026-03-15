package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHKM_ListRegistry(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/registry")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp struct {
		Count int `json:"count"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Count != 0 {
		t.Errorf("count = %d, want 0", resp.Count)
	}
}
