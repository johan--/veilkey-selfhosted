package api

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"veilkey-keycenter/internal/crypto"
	"veilkey-keycenter/internal/db"
)

func TestInstallGateBlocksOperationalHKMAPIUntilInstallCompletes(t *testing.T) {
	srv, handler := setupHKMServer(t)
	if err := srv.db.SaveInstallSession(&db.InstallSession{
		SessionID:           crypto.GenerateUUID(),
		Version:             1,
		Language:            "en",
		Flow:                "quickstart",
		DeploymentMode:      "host-service",
		InstallScope:        "host-only",
		BootstrapMode:       "email",
		MailTransport:       "smtp",
		PlannedStagesJSON:   `["language","bootstrap","final_smoke"]`,
		CompletedStagesJSON: `["language","bootstrap"]`,
		LastStage:           "bootstrap",
	}); err != nil {
		t.Fatalf("SaveInstallSession: %v", err)
	}

	w := getJSON(handler, "/api/registry")
	if w.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected 503, got %d: %s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); body == "" || !containsAllParts(body, "install flow is not complete", "final_smoke") {
		t.Fatalf("expected install guidance in body, got %q", body)
	}
}

func TestInstallGateAllowsOperationalHKMAPIAfterInstallCompletes(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/registry")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestStatusIncludesInstallState(t *testing.T) {
	srv, handler := setupHKMServer(t)
	if err := srv.db.SaveInstallSession(&db.InstallSession{
		SessionID:           "install-session-1",
		Version:             1,
		Language:            "en",
		Flow:                "advanced",
		DeploymentMode:      "host-service",
		InstallScope:        "host-only",
		BootstrapMode:       "email",
		MailTransport:       "smtp",
		PlannedStagesJSON:   `["language","deployment_mode","verification_and_resume"]`,
		CompletedStagesJSON: `["language","deployment_mode"]`,
		LastStage:           "deployment_mode",
	}); err != nil {
		t.Fatalf("SaveInstallSession: %v", err)
	}

	w := getJSON(handler, "/api/status")
	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Install installAccessState `json:"install"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal status response: %v", err)
	}
	if !resp.Install.Exists {
		t.Fatalf("expected install.exists=true")
	}
	if resp.Install.Complete {
		t.Fatalf("expected install.complete=false")
	}
	if resp.Install.SessionID != "install-session-1" {
		t.Fatalf("expected install.session_id=install-session-1, got %q", resp.Install.SessionID)
	}
	if resp.Install.FinalStage != "verification_and_resume" {
		t.Fatalf("expected install.final_stage=verification_and_resume, got %q", resp.Install.FinalStage)
	}
	if resp.Install.LastStage != "deployment_mode" {
		t.Fatalf("expected install.last_stage=deployment_mode, got %q", resp.Install.LastStage)
	}
}

func containsAllParts(s string, parts ...string) bool {
	for _, part := range parts {
		if !strings.Contains(s, part) {
			return false
		}
	}
	return true
}
