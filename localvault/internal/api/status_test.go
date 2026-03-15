package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"veilkey-localvault/internal/crypto"
	"veilkey-localvault/internal/db"
)

func setupStatusTestServer(t *testing.T) *Server {
	t.Helper()

	database, err := db.New(filepath.Join(t.TempDir(), "localvault.db"))
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}

	kek := []byte("0123456789abcdef0123456789abcdef")
	dek := []byte("abcdef0123456789abcdef0123456789")
	encDEK, encNonce, err := crypto.Encrypt(kek, dek)
	if err != nil {
		t.Fatalf("Encrypt DEK: %v", err)
	}
	if err := database.SaveNodeInfo(&db.NodeInfo{
		NodeID:   "93a8094e-ad3f-4143-b3f3-8551275f24a7",
		DEK:      encDEK,
		DEKNonce: encNonce,
		Version:  1,
	}); err != nil {
		t.Fatalf("SaveNodeInfo: %v", err)
	}
	if err := database.SaveConfig("VAULT_HASH", "93a8094e"); err != nil {
		t.Fatalf("SaveConfig VAULT_HASH: %v", err)
	}
	if err := database.SaveConfig("VAULT_NAME", "proxmox-test-lab-veilkey"); err != nil {
		t.Fatalf("SaveConfig VAULT_NAME: %v", err)
	}

	server := NewServer(database, kek, []string{"127.0.0.1"})
	server.SetIdentity(&NodeIdentity{
		NodeID:    "93a8094e-ad3f-4143-b3f3-8551275f24a7",
		Version:   1,
		VaultHash: "93a8094e",
		VaultName: "proxmox-test-lab-veilkey",
	})
	return server
}

func TestHandleStatusIncludesVaultIdentity(t *testing.T) {
	server := setupStatusTestServer(t)
	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	w := httptest.NewRecorder()

	server.handleStatus(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp map[string]any
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["vault_hash"] != "93a8094e" {
		t.Fatalf("vault_hash = %#v", resp["vault_hash"])
	}
	if resp["vault_name"] != "proxmox-test-lab-veilkey" {
		t.Fatalf("vault_name = %#v", resp["vault_name"])
	}
	if resp["vault_id"] != "proxmox-test-lab-veilkey:93a8094e" {
		t.Fatalf("vault_id = %#v", resp["vault_id"])
	}
	if resp["vault_node_uuid"] != "93a8094e-ad3f-4143-b3f3-8551275f24a7" {
		t.Fatalf("vault_node_uuid = %#v", resp["vault_node_uuid"])
	}
	if resp["node_id"] != "93a8094e-ad3f-4143-b3f3-8551275f24a7" {
		t.Fatalf("node_id = %#v", resp["node_id"])
	}
	if resp["mode"] != "vault" {
		t.Fatalf("mode = %#v", resp["mode"])
	}
}
