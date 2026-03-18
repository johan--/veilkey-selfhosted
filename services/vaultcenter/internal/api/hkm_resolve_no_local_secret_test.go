package api

import (
	"net/http"
	"testing"
	"veilkey-vaultcenter/internal/crypto"
	"veilkey-vaultcenter/internal/db"
)

func TestResolveSecretDoesNotUseLocalSecretTable(t *testing.T) {
	srv, handler := setupTestServer(t)

	info, err := srv.db.GetNodeInfo()
	if err != nil {
		t.Fatalf("GetNodeInfo: %v", err)
	}
	nodeDEK, err := crypto.DecryptDEK(srv.kek, info.DEK, info.DEKNonce)
	if err != nil {
		t.Fatalf("DecryptDEK: %v", err)
	}
	ciphertext, nonce, err := crypto.Encrypt(nodeDEK, []byte("legacy-local-secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if err := srv.db.SaveSecret(&db.Secret{
		ID:         crypto.GenerateUUID(),
		Name:       "LOCAL_ONLY_SECRET",
		Ref:        "deadbeef",
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Version:    info.Version,
	}); err != nil {
		t.Fatalf("SaveSecret: %v", err)
	}

	w := getJSON(handler, "/api/resolve/VK:LOCAL:deadbeef")
	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 when only local secret row exists, got %d body=%s", w.Code, w.Body.String())
	}
}
