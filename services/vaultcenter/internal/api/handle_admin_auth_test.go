package api

import (
	"os"
	"strings"
	"testing"
)

func TestAdminChangePasswordRequiresOwnerPassword(t *testing.T) {
	// The change-password handler must ONLY accept owner_password for verification.
	// It must NOT accept admin session, current admin password, or any other auth method.
	src, err := os.ReadFile("handle_admin_auth.go")
	if err != nil {
		t.Fatalf("failed to read handle_admin_auth.go: %v", err)
	}
	code := string(src)

	// Must have the handler
	if !strings.Contains(code, "handleAdminChangePassword") {
		t.Fatal("handleAdminChangePassword handler must exist")
	}

	// Must verify owner password via KEK derivation
	if !strings.Contains(code, "crypto.DeriveKEK(req.OwnerPassword") {
		t.Error("handleAdminChangePassword must verify owner password via crypto.DeriveKEK")
	}

	// Must decrypt DEK to prove owner password is correct
	if !strings.Contains(code, "crypto.Decrypt(kek, info.DEK, info.DEKNonce)") {
		t.Error("handleAdminChangePassword must verify KEK by decrypting DEK")
	}

	// The handler function body must NOT call resolveAdminSession
	// Extract the handler function body
	idx := strings.Index(code, "func (s *Server) handleAdminChangePassword")
	if idx >= 0 {
		// Find the next func definition to bound the search
		rest := code[idx:]
		nextFunc := strings.Index(rest[1:], "\nfunc ")
		if nextFunc > 0 {
			handlerBody := rest[:nextFunc+1]
			if strings.Contains(handlerBody, "resolveAdminSession") {
				t.Error("handleAdminChangePassword must NOT use admin session — only owner password")
			}
		}
	}
}

func TestAdminChangePasswordRouteHasTrustedIPGuard(t *testing.T) {
	src, err := os.ReadFile("handlers.go")
	if err != nil {
		t.Fatalf("failed to read handlers.go: %v", err)
	}
	code := string(src)

	// The change-password route must be behind requireTrustedIP
	if !strings.Contains(code, "change-password") {
		t.Fatal("change-password route must be registered in handlers.go")
	}
	if !strings.Contains(code, "requireTrustedIP(s.handleAdminChangePassword)") {
		t.Error("change-password route must be wrapped with requireTrustedIP")
	}
	// Must also require unlock
	if !strings.Contains(code, "requireUnlocked") {
		t.Error("change-password route must require server to be unlocked")
	}
}

func TestNoDirectDBAdminPasswordModification(t *testing.T) {
	// Ensure no handler directly modifies admin_auth_configs table
	// except through db.SetAdminPassword (which uses bcrypt)
	src, err := os.ReadFile("handle_admin_auth.go")
	if err != nil {
		t.Fatalf("failed to read handle_admin_auth.go: %v", err)
	}
	code := string(src)

	// Must use db.SetAdminPassword, not direct SQL
	if strings.Contains(code, "password_hash") && strings.Contains(code, "UPDATE") {
		t.Error("must not directly modify password_hash — use db.SetAdminPassword")
	}
	if strings.Contains(code, "admin_auth_configs") {
		t.Error("must not reference admin_auth_configs table directly — use db methods")
	}
}
