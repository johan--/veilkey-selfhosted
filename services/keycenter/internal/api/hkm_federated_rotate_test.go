package api

import (
	"encoding/json"
	"net/http"
	"testing"
	"veilkey-keycenter/internal/crypto"
	"veilkey-keycenter/internal/db"
)

func TestHKM_FederatedRotate(t *testing.T) {
	srv, handler := setupHKMServer(t)

	mockChild := newMockRekeyServer()
	defer mockChild.Close()

	childNodeID := crypto.GenerateUUID()
	childDEK, _ := crypto.GenerateKey()
	parentDEK, _ := srv.getLocalDEK()
	encChildDEK, childNonce, _ := crypto.Encrypt(parentDEK, childDEK)

	srv.db.RegisterChild(&db.Child{
		NodeID: childNodeID, Label: "rotate-child", URL: mockChild.URL,
		EncryptedDEK: encChildDEK, Nonce: childNonce, Version: 1,
	})

	w := postJSON(handler, "/api/federation/rotate", nil)
	if w.Code != http.StatusOK {
		t.Fatalf("rotate: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Success int `json:"success"`
		Failed  int `json:"failed"`
		Total   int `json:"total"`
	}
	json.Unmarshal(w.Body.Bytes(), &resp)
	if resp.Success != 1 {
		t.Errorf("success = %d, want 1", resp.Success)
	}
	if resp.Total != 1 {
		t.Errorf("total = %d, want 1", resp.Total)
	}
}
