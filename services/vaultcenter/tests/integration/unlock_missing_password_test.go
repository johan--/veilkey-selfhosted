package integration_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntegration_UnlockMissingPassword(t *testing.T) {
	_, handler, _ := setupServerWithPassword(t, "some-password")

	w := postJSON(handler, "/api/unlock", map[string]string{"password": ""})
	if w.Code != http.StatusBadRequest {
		t.Errorf("empty password: expected 400, got %d", w.Code)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/unlock", bytes.NewReader([]byte("not-json")))
	req.Header.Set("Content-Type", "application/json")
	rw := httptest.NewRecorder()
	handler.ServeHTTP(rw, req)
	if rw.Code != http.StatusBadRequest {
		t.Errorf("malformed body: expected 400, got %d", rw.Code)
	}
}
