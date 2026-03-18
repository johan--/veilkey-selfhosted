package integration_test

import (
	"net/http"
	"testing"
)

func TestIntegration_TrustedIP_EmptyList(t *testing.T) {
	_, handler := setupTestServer(t)

	for _, ip := range []string{"1.2.3.4", "192.168.99.1", "10.0.0.1"} {
		w := getJSONFromIP(handler, "/api/status", ip+":1234")
		if w.Code != http.StatusOK {
			t.Errorf("IP %s: expected 200, got %d", ip, w.Code)
		}
	}
}
