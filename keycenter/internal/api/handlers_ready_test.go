package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestReadyEndpoint(t *testing.T) {
	_, handler := setupTestServer(t)

	w := getJSON(handler, "/ready")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp["status"] != "ready" {
		t.Errorf("status = %q, want ready", resp["status"])
	}
}
