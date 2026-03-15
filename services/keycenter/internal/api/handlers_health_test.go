package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHealthEndpoint(t *testing.T) {
	_, handler := setupTestServer(t)

	w := getJSON(handler, "/health")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
	var resp map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("parse: %v", err)
	}
	if resp["status"] != "ok" {
		t.Errorf("status = %q, want ok", resp["status"])
	}
}
