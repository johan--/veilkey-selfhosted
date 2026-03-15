package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestKeycenterDoesNotExposeDirectSecretCRUDRoutes(t *testing.T) {
	_, handler := setupHKMServer(t)

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{name: "list", method: http.MethodGet, path: "/api/secrets"},
		{name: "save", method: http.MethodPost, path: "/api/secrets"},
		{name: "get", method: http.MethodGet, path: "/api/secrets/TEST_SECRET"},
		{name: "delete", method: http.MethodDelete, path: "/api/secrets/TEST_SECRET"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(`{"name":"TEST_SECRET","value":"secret-value"}`))
			w := httptest.NewRecorder()
			handler.ServeHTTP(w, req)
			if w.Code != http.StatusNotFound {
				t.Fatalf("%s %s => %d, want 404", tt.method, tt.path, w.Code)
			}
		})
	}
}
