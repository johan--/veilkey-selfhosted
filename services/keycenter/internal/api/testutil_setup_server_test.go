package api

import (
	"net/http"
	"path/filepath"
	"testing"
	"veilkey-keycenter/internal/crypto"
	"veilkey-keycenter/internal/db"
)

func setupTestServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	dir := t.TempDir()
	database, err := db.New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	kek, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey (KEK): %v", err)
	}
	dek, err := crypto.GenerateKey()
	if err != nil {
		t.Fatalf("GenerateKey (DEK): %v", err)
	}
	encDEK, nonce, err := crypto.EncryptDEK(kek, dek)
	if err != nil {
		t.Fatalf("EncryptDEK: %v", err)
	}
	if err = database.SaveNodeInfo(&db.NodeInfo{
		NodeID:   crypto.GenerateUUID(),
		DEK:      encDEK,
		DEKNonce: nonce,
		Version:  1,
	}); err != nil {
		t.Fatalf("SaveNodeInfo: %v", err)
	}

	srv := NewServer(database, kek, []string{})
	if err := database.SaveInstallSession(&db.InstallSession{
		SessionID:           crypto.GenerateUUID(),
		Version:             1,
		Language:            "en",
		Flow:                "quickstart",
		DeploymentMode:      "host-service",
		InstallScope:        "host-only",
		BootstrapMode:       "email",
		MailTransport:       "smtp",
		PlannedStagesJSON:   `["language","bootstrap","final_smoke"]`,
		CompletedStagesJSON: `["language","bootstrap","final_smoke"]`,
		LastStage:           "final_smoke",
	}); err != nil {
		t.Fatalf("SaveInstallSession: %v", err)
	}
	handler := srv.SetupRoutes()
	return srv, handler
}
