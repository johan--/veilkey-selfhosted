package integration_test

import (
	"net/http"
	"path/filepath"
	"testing"
	"veilkey-keycenter/internal/api"
	"veilkey-keycenter/internal/crypto"
	"veilkey-keycenter/internal/db"
)

func setupTestServer(t *testing.T) (*api.Server, http.Handler) {
	t.Helper()
	database, err := db.New(filepath.Join(t.TempDir(), "test.db"))
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

	srv := api.NewServer(database, kek, []string{})
	handler := srv.SetupRoutes()
	return srv, handler
}
