package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenReadinessRehearsalStatusAndPlaintextBoundary(t *testing.T) {
	server := setupStatusTestServer(t)
	handler := server.SetupRoutes()

	statusReq := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	statusW := httptest.NewRecorder()
	handler.ServeHTTP(statusW, statusReq)
	if statusW.Code != http.StatusOK {
		t.Fatalf("status expected 200, got %d: %s", statusW.Code, statusW.Body.String())
	}
	if !bytes.Contains(statusW.Body.Bytes(), []byte(`"vault_id":"proxmox-test-lab-veilkey:93a8094e"`)) {
		t.Fatalf("status should include vault identity: %s", statusW.Body.String())
	}

	plaintextReq := httptest.NewRequest(http.MethodPost, "/api/encrypt", bytes.NewBufferString(`{"plaintext":"secret"}`))
	plaintextReq.RemoteAddr = "127.0.0.1:12345"
	plaintextW := httptest.NewRecorder()
	handler.ServeHTTP(plaintextW, plaintextReq)
	if plaintextW.Code != http.StatusForbidden {
		t.Fatalf("encrypt expected 403, got %d: %s", plaintextW.Code, plaintextW.Body.String())
	}
	if !bytes.Contains(plaintextW.Body.Bytes(), []byte(keycenterOnlyDecryptMessage)) {
		t.Fatalf("encrypt should mention keycenter-only restriction: %s", plaintextW.Body.String())
	}
}
