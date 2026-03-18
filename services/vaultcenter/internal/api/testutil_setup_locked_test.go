package api

import (
	"net/http"
	"path/filepath"
	"testing"
	"veilkey-vaultcenter/internal/crypto"
	"veilkey-vaultcenter/internal/db"
)

func setupLockedServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	dir := t.TempDir()
	database, err := db.New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	kek, _ := crypto.GenerateKey()
	dek, _ := crypto.GenerateKey()
	encDEK, nonce, _ := crypto.EncryptDEK(kek, dek)
	if err := database.SaveNodeInfo(&db.NodeInfo{
		NodeID:   crypto.GenerateUUID(),
		DEK:      encDEK,
		DEKNonce: nonce,
		Version:  1,
	}); err != nil {
		t.Fatalf("SaveNodeInfo: %v", err)
	}

	srv := NewServer(database, nil, []string{})
	handler := srv.SetupRoutes()
	return srv, handler
}
