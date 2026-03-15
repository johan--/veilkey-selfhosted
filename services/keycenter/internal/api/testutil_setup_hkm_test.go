package api

import (
	"net/http"
	"path/filepath"
	"testing"
	"veilkey-keycenter/internal/crypto"
	"veilkey-keycenter/internal/db"
)

func setupHKMServer(t *testing.T) (*Server, http.Handler) {
	t.Helper()
	dir := t.TempDir()
	database, err := db.New(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}
	t.Cleanup(func() { database.Close() })

	kek, _ := crypto.GenerateKey()
	dek, _ := crypto.GenerateKey()
	encDEK, encNonce, _ := crypto.Encrypt(kek, dek)

	nodeID := crypto.GenerateUUID()
	if err := database.SaveNodeInfo(&db.NodeInfo{
		NodeID:   nodeID,
		DEK:      encDEK,
		DEKNonce: encNonce,
		Version:  1,
	}); err != nil {
		t.Fatalf("SaveNodeInfo: %v", err)
	}

	srv := NewServer(database, kek, []string{})
	srv.SetIdentity(&NodeIdentity{
		NodeID:  nodeID,
		Version: 1,
		IsHKM:   true,
	})
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
