package integration_test

import (
	"net/http"
	"testing"
)

func TestIntegration_TrustedIP(t *testing.T) {
	_, handler := setupTrustedIPServer(t, []string{"10.0.0.1", "172.16.0.0/24"})

	t.Run("allowed exact IP", func(t *testing.T) {
		w := postJSONFromIP(handler, "/api/unlock", "10.0.0.1:12345", map[string]string{"password": "any"})
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("allowed CIDR IP", func(t *testing.T) {
		w := postJSONFromIP(handler, "/api/unlock", "172.16.0.5:9999", map[string]string{"password": "any"})
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("blocked IP", func(t *testing.T) {
		w := postJSONFromIP(handler, "/api/unlock", "192.168.1.99:12345", map[string]string{"password": "any"})
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})
	t.Run("health not restricted", func(t *testing.T) {
		w := getJSONFromIP(handler, "/health", "192.168.1.99:12345")
		if w.Code != http.StatusOK {
			t.Errorf("expected 200, got %d", w.Code)
		}
	})
	t.Run("unlock restricted", func(t *testing.T) {
		w := postJSONFromIP(handler, "/api/unlock", "192.168.1.99:12345", map[string]string{"password": "any"})
		if w.Code != http.StatusForbidden {
			t.Errorf("expected 403, got %d", w.Code)
		}
	})
}
