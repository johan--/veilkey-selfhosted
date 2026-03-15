package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestBlockedReadPathsReturnLocked(t *testing.T) {
	server := setupReencryptTestServer(t)
	handler := server.SetupRoutes()

	if err := server.db.UpdateSecretLifecycle("deadbeef", "LOCAL", "block"); err != nil {
		t.Fatalf("UpdateSecretLifecycle: %v", err)
	}
	if err := server.db.SaveConfig("BLOCKED_CFG", "value"); err != nil {
		t.Fatalf("SaveConfig: %v", err)
	}
	if err := server.db.UpdateConfigLifecycle("BLOCKED_CFG", "LOCAL", "block"); err != nil {
		t.Fatalf("UpdateConfigLifecycle: %v", err)
	}

	tests := []struct {
		name string
		path string
		want string
	}{
		{name: "cipher", path: "/api/cipher/deadbeef", want: "VK:LOCAL:deadbeef"},
		{name: "config", path: "/api/configs/BLOCKED_CFG", want: "VE:LOCAL:BLOCKED_CFG"},
		{name: "archive", path: "/api/archive", want: "VK:LOCAL:deadbeef"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			switch tt.name {
			case "archive":
				r := httptest.NewRequest(http.MethodPost, tt.path, strings.NewReader(`{"ciphertext":"VK:LOCAL:deadbeef"}`))
				handler.ServeHTTP(w, r)
			default:
				r := httptest.NewRequest(http.MethodGet, tt.path, nil)
				if tt.name == "cipher" {
					r.RemoteAddr = "127.0.0.1:12345"
				}
				handler.ServeHTTP(w, r)
			}
			if w.Code != http.StatusLocked {
				t.Fatalf("%s => %d, want 423: %s", tt.path, w.Code, w.Body.String())
			}
			if !strings.Contains(w.Body.String(), tt.want) {
				t.Fatalf("body %q does not contain %q", w.Body.String(), tt.want)
			}
		})
	}
}
