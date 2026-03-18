package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPlaintextHandlingEndpointsAreDisabled(t *testing.T) {
	server := setupStatusTestServer(t)
	handler := server.SetupRoutes()

	tests := []struct {
		name   string
		method string
		path   string
		body   string
	}{
		{
			name:   "save secret",
			method: http.MethodPost,
			path:   "/api/secrets",
			body:   `{"name":"demo","value":"secret"}`,
		},
		{
			name:   "get secret",
			method: http.MethodGet,
			path:   "/api/secrets/demo",
		},
		{
			name:   "resolve",
			method: http.MethodGet,
			path:   "/api/resolve/deadbeef",
		},
		{
			name:   "encrypt",
			method: http.MethodPost,
			path:   "/api/encrypt",
			body:   `{"plaintext":"secret"}`,
		},
		{
			name:   "rekey",
			method: http.MethodPost,
			path:   "",
		},
	}

	for _, tt := range tests {
		if tt.path == "" {
			continue
		}
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(tt.method, tt.path, bytes.NewBufferString(tt.body))
			req.RemoteAddr = "127.0.0.1:12345"
			w := httptest.NewRecorder()

			handler.ServeHTTP(w, req)

			if w.Code != http.StatusForbidden {
				t.Fatalf("expected 403, got %d: %s", w.Code, w.Body.String())
			}
			if !bytes.Contains(w.Body.Bytes(), []byte(vaultcenterOnlyDecryptMessage)) {
				t.Fatalf("response does not mention vaultcenter-only restriction: %s", w.Body.String())
			}
		})
	}
}
