package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"veilkey-keycenter/internal/db"
)

func TestAdminTOTPEnrollmentAndSessionFlow(t *testing.T) {
	_, handler := setupTestServer(t)

	start := postJSON(handler, "/api/admin/auth/totp/enroll/start", map[string]any{})
	if start.Code != http.StatusOK {
		t.Fatalf("enroll start: expected 200, got %d: %s", start.Code, start.Body.String())
	}
	var startResp struct {
		Secret   string                    `json:"secret"`
		Otpauth  string                    `json:"otpauth_uri"`
		Settings adminAuthSettingsResponse `json:"settings"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &startResp); err != nil {
		t.Fatalf("decode start response: %v", err)
	}
	if startResp.Secret == "" || !strings.Contains(startResp.Otpauth, "otpauth://totp/") {
		t.Fatalf("expected secret and otpauth uri, got %+v", startResp)
	}
	if !startResp.Settings.PendingEnrollment {
		t.Fatalf("expected pending enrollment")
	}

	code := totpCode(startResp.Secret, time.Now().UTC())
	verify := postJSON(handler, "/api/admin/auth/totp/enroll/verify", map[string]any{"code": code})
	if verify.Code != http.StatusOK {
		t.Fatalf("enroll verify: expected 200, got %d: %s", verify.Code, verify.Body.String())
	}

	settings := getJSON(handler, "/api/admin/auth/settings")
	if settings.Code != http.StatusOK {
		t.Fatalf("settings: expected 200, got %d: %s", settings.Code, settings.Body.String())
	}
	var settingsResp adminAuthSettingsResponse
	if err := json.Unmarshal(settings.Body.Bytes(), &settingsResp); err != nil {
		t.Fatalf("decode settings response: %v", err)
	}
	if !settingsResp.TOTPEnabled || !settingsResp.TOTPEnrolled {
		t.Fatalf("expected totp enabled and enrolled, got %+v", settingsResp)
	}

	login := postJSON(handler, "/api/admin/session/login", map[string]any{"code": totpCode(startResp.Secret, time.Now().UTC())})
	if login.Code != http.StatusOK {
		t.Fatalf("login: expected 200, got %d: %s", login.Code, login.Body.String())
	}
	cookies := login.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected admin session cookie")
	}

	req := httptest.NewRequest(http.MethodGet, "/api/admin/session", nil)
	req.AddCookie(cookies[0])
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("session get: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAdminRevealAuthorizeWritesAuditAndUpdatesCatalog(t *testing.T) {
	srv, handler := setupTestServer(t)
	secret := enrollAndLoginAdmin(t, handler)

	entry := &db.SecretCatalog{
		SecretCanonicalID: "vh-a:API_KEY",
		SecretName:        "API_KEY",
		DisplayName:       "API Key",
		Class:             "key",
		Scope:             "LOCAL",
		Status:            "active",
		VaultNodeUUID:     "node-a",
		VaultRuntimeHash:  "runtime-a",
		VaultHash:         "vh-a",
		RefCanonical:      "VK:LOCAL:deadbeef",
		FieldsPresentJSON: `["VALUE"]`,
	}
	if err := srv.db.SaveSecretCatalog(entry); err != nil {
		t.Fatalf("SaveSecretCatalog: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/admin/reveal-authorize", strings.NewReader(`{"ref":"VK:LOCAL:deadbeef","reason":"incident review"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(secret)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("reveal authorize: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	updated, err := srv.db.GetSecretCatalogByRef("VK:LOCAL:deadbeef")
	if err != nil {
		t.Fatalf("GetSecretCatalogByRef: %v", err)
	}
	if updated.LastRevealedAt == nil {
		t.Fatalf("expected last_revealed_at to be updated")
	}

	rows, err := srv.db.ListAuditEvents("secret", "VK:LOCAL:deadbeef")
	if err != nil {
		t.Fatalf("ListAuditEvents: %v", err)
	}
	if len(rows) == 0 || rows[0].Action != "admin_reveal_authorize" {
		t.Fatalf("expected admin_reveal_authorize audit row, got %+v", rows)
	}
}

func TestAdminRevealReturnsPlaintextWithinAuthorizedWindow(t *testing.T) {
	srv, handler := setupHKMServer(t)
	cookie := enrollAndLoginAdmin(t, handler)

	_, agentHash := registerMockAgent(t, srv, "admin-reveal-agent", map[string]string{}, nil)
	save := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "OPENAI_API_KEY",
		"value": "super-secret-123",
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save secret: expected 200, got %d: %s", save.Code, save.Body.String())
	}
	var saveResp struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(save.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("decode save response: %v", err)
	}

	authReq := httptest.NewRequest(http.MethodPost, "/api/admin/reveal-authorize", strings.NewReader(`{"ref":"VK:TEMP:`+saveResp.Ref+`","reason":"incident review"}`))
	authReq.Header.Set("Content-Type", "application/json")
	authReq.AddCookie(cookie)
	authW := httptest.NewRecorder()
	handler.ServeHTTP(authW, authReq)
	if authW.Code != http.StatusOK {
		t.Fatalf("reveal authorize: expected 200, got %d: %s", authW.Code, authW.Body.String())
	}

	revealReq := httptest.NewRequest(http.MethodPost, "/api/admin/reveal", strings.NewReader(`{"ref":"VK:TEMP:`+saveResp.Ref+`"}`))
	revealReq.Header.Set("Content-Type", "application/json")
	revealReq.AddCookie(cookie)
	revealW := httptest.NewRecorder()
	handler.ServeHTTP(revealW, revealReq)
	if revealW.Code != http.StatusOK {
		t.Fatalf("reveal: expected 200, got %d: %s", revealW.Code, revealW.Body.String())
	}
	var revealResp struct {
		Ref   string `json:"ref"`
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(revealW.Body.Bytes(), &revealResp); err != nil {
		t.Fatalf("decode reveal response: %v", err)
	}
	if revealResp.Ref != "VK:TEMP:"+saveResp.Ref || revealResp.Name != "OPENAI_API_KEY" || revealResp.Value != "super-secret-123" {
		t.Fatalf("unexpected reveal payload: %+v", revealResp)
	}

	rows, err := srv.db.ListAuditEvents("secret", "VK:TEMP:"+saveResp.Ref)
	if err != nil {
		t.Fatalf("ListAuditEvents: %v", err)
	}
	found := false
	for _, row := range rows {
		if row.Action == "admin_reveal" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected admin_reveal audit row, got %+v", rows)
	}
}

func TestAdminRevealRejectsWithoutAuthorizedWindow(t *testing.T) {
	srv, handler := setupHKMServer(t)
	cookie := enrollAndLoginAdmin(t, handler)

	_, agentHash := registerMockAgent(t, srv, "admin-reveal-forbidden-agent", map[string]string{}, nil)
	save := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "OPENAI_API_KEY",
		"value": "super-secret-123",
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save secret: expected 200, got %d: %s", save.Code, save.Body.String())
	}
	var saveResp struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(save.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("decode save response: %v", err)
	}

	revealReq := httptest.NewRequest(http.MethodPost, "/api/admin/reveal", strings.NewReader(`{"ref":"VK:TEMP:`+saveResp.Ref+`"}`))
	revealReq.Header.Set("Content-Type", "application/json")
	revealReq.AddCookie(cookie)
	revealW := httptest.NewRecorder()
	handler.ServeHTTP(revealW, revealReq)
	if revealW.Code != http.StatusForbidden {
		t.Fatalf("reveal without auth window: expected 403, got %d: %s", revealW.Code, revealW.Body.String())
	}
}

func TestAdminRebindApprovalsListAndApprove(t *testing.T) {
	srv, handler := setupHKMServer(t)
	cookie := enrollAndLoginAdmin(t, handler)

	_, agentHash := registerMockAgent(t, srv, "admin-approve-rebind-agent", map[string]string{}, map[string]string{
		"API_KEY": "secret-value",
	})
	agent, err := srv.db.GetAgentByHash(agentHash)
	if err != nil {
		t.Fatalf("GetAgentByHash: %v", err)
	}
	if _, err := srv.db.AdvanceAgentRebind(agent.NodeID, "key_version_mismatch", time.Now().UTC()); err != nil {
		t.Fatalf("AdvanceAgentRebind: %v", err)
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/approvals/rebind", nil)
	listReq.AddCookie(cookie)
	listW := httptest.NewRecorder()
	handler.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("approvals list: expected 200, got %d: %s", listW.Code, listW.Body.String())
	}
	var listResp struct {
		Count     int `json:"count"`
		Approvals []struct {
			VaultRuntimeHash  string `json:"vault_runtime_hash"`
			RebindRequired    bool   `json:"rebind_required"`
			EvidenceReady     bool   `json:"evidence_ready"`
			LatestSecureInput *struct {
				Token  string `json:"token"`
				Status string `json:"status"`
			} `json:"latest_secure_input"`
		} `json:"approvals"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode approvals list: %v", err)
	}
	if listResp.Count == 0 || len(listResp.Approvals) == 0 || listResp.Approvals[0].VaultRuntimeHash != agentHash || !listResp.Approvals[0].RebindRequired || listResp.Approvals[0].EvidenceReady || listResp.Approvals[0].LatestSecureInput != nil {
		t.Fatalf("unexpected approvals list: %+v", listResp)
	}

	planReq := httptest.NewRequest(http.MethodGet, "/api/admin/approvals/rebind/"+agentHash, nil)
	planReq.AddCookie(cookie)
	planW := httptest.NewRecorder()
	handler.ServeHTTP(planW, planReq)
	if planW.Code != http.StatusOK {
		t.Fatalf("rebind plan: expected 200, got %d: %s", planW.Code, planW.Body.String())
	}
	var planResp struct {
		Status              string `json:"status"`
		CurrentKeyVersion   int    `json:"current_key_version"`
		NextKeyVersion      int    `json:"next_key_version"`
		SecureInputRequired bool   `json:"secure_input_required"`
	}
	if err := json.Unmarshal(planW.Body.Bytes(), &planResp); err != nil {
		t.Fatalf("decode plan response: %v", err)
	}
	if planResp.Status != "plan" || planResp.CurrentKeyVersion != 1 || planResp.NextKeyVersion != 2 || !planResp.SecureInputRequired {
		t.Fatalf("unexpected plan response: %+v", planResp)
	}

	evidenceToken := createSubmittedSecureInputEvidence(t, handler, cookie, agentHash)

	listReqAfterEvidence := httptest.NewRequest(http.MethodGet, "/api/admin/approvals/rebind", nil)
	listReqAfterEvidence.AddCookie(cookie)
	listWAfterEvidence := httptest.NewRecorder()
	handler.ServeHTTP(listWAfterEvidence, listReqAfterEvidence)
	if listWAfterEvidence.Code != http.StatusOK {
		t.Fatalf("approvals list after evidence: expected 200, got %d: %s", listWAfterEvidence.Code, listWAfterEvidence.Body.String())
	}
	var listRespAfterEvidence struct {
		Approvals []struct {
			VaultRuntimeHash  string `json:"vault_runtime_hash"`
			EvidenceReady     bool   `json:"evidence_ready"`
			LatestSecureInput *struct {
				Token  string `json:"token"`
				Status string `json:"status"`
			} `json:"latest_secure_input"`
		} `json:"approvals"`
	}
	if err := json.Unmarshal(listWAfterEvidence.Body.Bytes(), &listRespAfterEvidence); err != nil {
		t.Fatalf("decode approvals list after evidence: %v", err)
	}
	if len(listRespAfterEvidence.Approvals) == 0 || listRespAfterEvidence.Approvals[0].VaultRuntimeHash != agentHash || !listRespAfterEvidence.Approvals[0].EvidenceReady || listRespAfterEvidence.Approvals[0].LatestSecureInput == nil || listRespAfterEvidence.Approvals[0].LatestSecureInput.Token != evidenceToken || listRespAfterEvidence.Approvals[0].LatestSecureInput.Status != "submitted" {
		t.Fatalf("unexpected approvals list after evidence: %+v", listRespAfterEvidence)
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/api/admin/approvals/rebind/"+agentHash+"/approve", strings.NewReader(`{"confirm":"APPROVE `+strings.ToUpper(agentHash)+`","reason":"rotation mismatch reviewed"}`))
	approveReq.Header.Set("Content-Type", "application/json")
	approveReq.AddCookie(cookie)
	approveW := httptest.NewRecorder()
	handler.ServeHTTP(approveW, approveReq)
	if approveW.Code != http.StatusOK {
		t.Fatalf("approve rebind: expected 200, got %d: %s", approveW.Code, approveW.Body.String())
	}
	var approveResp struct {
		Status           string `json:"status"`
		VaultRuntimeHash string `json:"vault_runtime_hash"`
		KeyVersion       int    `json:"key_version"`
		EvidenceToken    string `json:"evidence_token"`
	}
	if err := json.Unmarshal(approveW.Body.Bytes(), &approveResp); err != nil {
		t.Fatalf("decode approve response: %v", err)
	}
	if approveResp.Status != "approved" || approveResp.KeyVersion != 2 || approveResp.VaultRuntimeHash == "" || approveResp.VaultRuntimeHash == agentHash || approveResp.EvidenceToken != evidenceToken {
		t.Fatalf("unexpected approve response: %+v", approveResp)
	}

	updated, err := srv.db.GetAgentByNodeID(agent.NodeID)
	if err != nil {
		t.Fatalf("GetAgentByNodeID: %v", err)
	}
	if updated.RebindRequired || updated.BlockedAt != nil {
		t.Fatalf("expected rebind state to clear: %+v", updated)
	}

	rows, err := srv.db.ListAuditEvents("vault", agent.NodeID)
	if err != nil {
		t.Fatalf("ListAuditEvents: %v", err)
	}
	found := false
	for _, row := range rows {
		if row.Action == "admin_approve_rebind" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected admin_approve_rebind audit row, got %+v", rows)
	}
}

func TestAdminApproveRebindRejectsWithoutConfirmation(t *testing.T) {
	srv, handler := setupHKMServer(t)
	cookie := enrollAndLoginAdmin(t, handler)

	_, agentHash := registerMockAgent(t, srv, "admin-approve-rebind-confirm-agent", map[string]string{}, map[string]string{
		"API_KEY": "secret-value",
	})
	agent, err := srv.db.GetAgentByHash(agentHash)
	if err != nil {
		t.Fatalf("GetAgentByHash: %v", err)
	}
	if _, err := srv.db.AdvanceAgentRebind(agent.NodeID, "key_version_mismatch", time.Now().UTC()); err != nil {
		t.Fatalf("AdvanceAgentRebind: %v", err)
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/api/admin/approvals/rebind/"+agentHash+"/approve", strings.NewReader(`{"confirm":"APPROVE nope","reason":"checked"}`))
	approveReq.Header.Set("Content-Type", "application/json")
	approveReq.AddCookie(cookie)
	approveW := httptest.NewRecorder()
	handler.ServeHTTP(approveW, approveReq)
	if approveW.Code != http.StatusBadRequest {
		t.Fatalf("approve rebind confirmation mismatch: expected 400, got %d: %s", approveW.Code, approveW.Body.String())
	}
}

func TestAdminApproveRebindRejectsWithoutSubmittedEvidence(t *testing.T) {
	srv, handler := setupHKMServer(t)
	cookie := enrollAndLoginAdmin(t, handler)

	_, agentHash := registerMockAgent(t, srv, "admin-approve-rebind-evidence-agent", map[string]string{}, map[string]string{
		"API_KEY": "secret-value",
	})
	agent, err := srv.db.GetAgentByHash(agentHash)
	if err != nil {
		t.Fatalf("GetAgentByHash: %v", err)
	}
	if _, err := srv.db.AdvanceAgentRebind(agent.NodeID, "key_version_mismatch", time.Now().UTC()); err != nil {
		t.Fatalf("AdvanceAgentRebind: %v", err)
	}

	approveReq := httptest.NewRequest(http.MethodPost, "/api/admin/approvals/rebind/"+agentHash+"/approve", strings.NewReader(`{"confirm":"APPROVE `+strings.ToUpper(agentHash)+`","reason":"checked"}`))
	approveReq.Header.Set("Content-Type", "application/json")
	approveReq.AddCookie(cookie)
	approveW := httptest.NewRecorder()
	handler.ServeHTTP(approveW, approveReq)
	if approveW.Code != http.StatusConflict {
		t.Fatalf("approve rebind without evidence: expected 409, got %d: %s", approveW.Code, approveW.Body.String())
	}
}

func TestAdminScheduleAllRotationsAndCleanup(t *testing.T) {
	srv, handler := setupHKMServer(t)
	cookie := enrollAndLoginAdmin(t, handler)

	registerMockAgent(t, srv, "admin-rotate-a", map[string]string{}, nil)
	registerMockAgent(t, srv, "admin-rotate-b", map[string]string{}, nil)

	rotateReq := httptest.NewRequest(http.MethodPost, "/api/admin/rotations/schedule-all", strings.NewReader(`{"confirm":"ROTATE ALL","reason":"quarterly rotation"}`))
	rotateReq.Header.Set("Content-Type", "application/json")
	rotateReq.AddCookie(cookie)
	rotateW := httptest.NewRecorder()
	handler.ServeHTTP(rotateW, rotateReq)
	if rotateW.Code != http.StatusOK {
		t.Fatalf("schedule rotations: expected 200, got %d: %s", rotateW.Code, rotateW.Body.String())
	}
	var rotateResp struct {
		Status string `json:"status"`
		Count  int    `json:"count"`
		Reason string `json:"reason"`
	}
	if err := json.Unmarshal(rotateW.Body.Bytes(), &rotateResp); err != nil {
		t.Fatalf("decode rotate response: %v", err)
	}
	if rotateResp.Status != "scheduled" || rotateResp.Count < 2 || rotateResp.Reason != "quarterly rotation" {
		t.Fatalf("unexpected rotate response: %+v", rotateResp)
	}

	if err := seedTrackedRefCleanupFixtures(t, srv, handler); err != nil {
		t.Fatalf("seed cleanup fixtures: %v", err)
	}

	previewReq := httptest.NewRequest(http.MethodPost, "/api/admin/tracked-refs/cleanup-preview", strings.NewReader(`{}`))
	previewReq.Header.Set("Content-Type", "application/json")
	previewReq.AddCookie(cookie)
	previewW := httptest.NewRecorder()
	handler.ServeHTTP(previewW, previewReq)
	if previewW.Code != http.StatusOK {
		t.Fatalf("cleanup preview: expected 200, got %d: %s", previewW.Code, previewW.Body.String())
	}
	var previewResp trackedRefCleanupResponse
	if err := json.Unmarshal(previewW.Body.Bytes(), &previewResp); err != nil {
		t.Fatalf("decode preview response: %v", err)
	}
	if previewResp.Status != "preview" || previewResp.Counts["actions"] == 0 {
		t.Fatalf("unexpected cleanup preview: %+v", previewResp)
	}

	applyReq := httptest.NewRequest(http.MethodPost, "/api/admin/tracked-refs/cleanup-apply", strings.NewReader(`{"confirm":"CLEANUP STALE REFS","reason":"ownership repair"}`))
	applyReq.Header.Set("Content-Type", "application/json")
	applyReq.AddCookie(cookie)
	applyW := httptest.NewRecorder()
	handler.ServeHTTP(applyW, applyReq)
	if applyW.Code != http.StatusOK {
		t.Fatalf("cleanup apply: expected 200, got %d: %s", applyW.Code, applyW.Body.String())
	}
	var applyResp trackedRefCleanupResponse
	if err := json.Unmarshal(applyW.Body.Bytes(), &applyResp); err != nil {
		t.Fatalf("decode apply response: %v", err)
	}
	if applyResp.Status != "applied" {
		t.Fatalf("unexpected cleanup apply: %+v", applyResp)
	}
}

func TestAdminScheduleSingleRotation(t *testing.T) {
	srv, handler := setupHKMServer(t)
	cookie := enrollAndLoginAdmin(t, handler)

	_, agentHash := registerMockAgent(t, srv, "admin-rotate-single", map[string]string{}, nil)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/rotations/"+agentHash+"/schedule", strings.NewReader(`{"confirm":"ROTATE `+strings.ToUpper(agentHash)+`","reason":"targeted rotation"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("schedule single rotation: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Status           string `json:"status"`
		RotationRequired bool   `json:"rotation_required"`
		KeyVersion       int    `json:"key_version"`
		VaultRuntimeHash string `json:"vault_runtime_hash"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode single rotation response: %v", err)
	}
	if resp.Status != "scheduled" || !resp.RotationRequired || resp.KeyVersion != 2 || resp.VaultRuntimeHash == "" {
		t.Fatalf("unexpected single rotation response: %+v", resp)
	}

	updated, err := srv.db.GetAgentByHash(resp.VaultRuntimeHash)
	if err != nil {
		t.Fatalf("GetAgentByHash(updated): %v", err)
	}
	if !updated.RotationRequired || updated.KeyVersion != 2 {
		t.Fatalf("expected rotation state to persist: %+v", updated)
	}
	rows, err := srv.db.ListAuditEvents("vault", updated.NodeID)
	if err != nil {
		t.Fatalf("ListAuditEvents: %v", err)
	}
	found := false
	for _, row := range rows {
		if row.Action == "admin_schedule_rotation_single" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected admin_schedule_rotation_single audit row, got %+v", rows)
	}
}

func TestAdminRecentAuditEndpointListsAdminActions(t *testing.T) {
	srv, handler := setupHKMServer(t)
	cookie := enrollAndLoginAdmin(t, handler)

	_, agentHash := registerMockAgent(t, srv, "admin-audit-recent", map[string]string{}, nil)
	req := httptest.NewRequest(http.MethodPost, "/api/admin/rotations/"+agentHash+"/schedule", strings.NewReader(`{"confirm":"ROTATE `+strings.ToUpper(agentHash)+`","reason":"recent audit check"}`))
	req.Header.Set("Content-Type", "application/json")
	req.AddCookie(cookie)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("schedule single rotation: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	listReq := httptest.NewRequest(http.MethodGet, "/api/admin/audit/recent?limit=5", nil)
	listReq.AddCookie(cookie)
	listW := httptest.NewRecorder()
	handler.ServeHTTP(listW, listReq)
	if listW.Code != http.StatusOK {
		t.Fatalf("recent admin audit: expected 200, got %d: %s", listW.Code, listW.Body.String())
	}
	var listResp struct {
		Count  int `json:"count"`
		Events []struct {
			Action string `json:"action"`
			Reason string `json:"reason"`
			Source string `json:"source"`
		} `json:"events"`
	}
	if err := json.Unmarshal(listW.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode recent admin audit: %v", err)
	}
	if listResp.Count == 0 || len(listResp.Events) == 0 {
		t.Fatalf("expected recent admin audit rows, got %+v", listResp)
	}
	found := false
	for _, event := range listResp.Events {
		if event.Action == "admin_schedule_rotation_single" && event.Reason == "recent audit check" && event.Source == "admin_rotate_single" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("expected recent admin audit to include single rotation event, got %+v", listResp.Events)
	}
}

func seedTrackedRefCleanupFixtures(t *testing.T, srv *Server, handler http.Handler) error {
	t.Helper()
	hbA := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"vault_node_uuid": "node-clean-a",
		"label":           "cleanup-a",
		"vault_hash":      "vh-clean-a",
		"vault_name":      "cleanup-a",
		"key_version":     1,
		"ip":              "127.0.0.1",
		"port":            19081,
		"secrets_count":   0,
		"configs_count":   0,
		"version":         1,
	})
	if hbA.Code != http.StatusOK {
		return fmt.Errorf("heartbeat cleanup-a: %s", hbA.Body.String())
	}
	hbB := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"vault_node_uuid": "node-clean-b",
		"label":           "cleanup-b",
		"vault_hash":      "vh-clean-b",
		"vault_name":      "cleanup-b",
		"key_version":     1,
		"ip":              "127.0.0.1",
		"port":            19082,
		"secrets_count":   0,
		"configs_count":   0,
		"version":         1,
	})
	if hbB.Code != http.StatusOK {
		return fmt.Errorf("heartbeat cleanup-b: %s", hbB.Body.String())
	}
	var respA, respB struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
	}
	if err := json.Unmarshal(hbA.Body.Bytes(), &respA); err != nil {
		return err
	}
	if err := json.Unmarshal(hbB.Body.Bytes(), &respB); err != nil {
		return err
	}
	for _, seed := range []struct {
		ref    string
		status string
		owner  string
	}{
		{"VK:TEMP:dup001", "temp", respA.VaultRuntimeHash},
		{"VK:LOCAL:dup001", "active", respA.VaultRuntimeHash},
		{"VE:LOCAL:APP_URL", "active", ""},
		{"VK:LOCAL:shared001", "active", respA.VaultRuntimeHash},
		{"VK:TEMP:shared001", "temp", respB.VaultRuntimeHash},
	} {
		if err := srv.upsertTrackedRef(seed.ref, 5, seed.status, seed.owner); err != nil {
			return err
		}
	}
	return nil
}

func enrollAndLoginAdmin(t *testing.T, handler http.Handler) *http.Cookie {
	t.Helper()
	start := postJSON(handler, "/api/admin/auth/totp/enroll/start", map[string]any{})
	if start.Code != http.StatusOK {
		t.Fatalf("enroll start: %d %s", start.Code, start.Body.String())
	}
	var startResp struct {
		Secret string `json:"secret"`
	}
	if err := json.Unmarshal(start.Body.Bytes(), &startResp); err != nil {
		t.Fatalf("decode start: %v", err)
	}
	verify := postJSON(handler, "/api/admin/auth/totp/enroll/verify", map[string]any{"code": totpCode(startResp.Secret, time.Now().UTC())})
	if verify.Code != http.StatusOK {
		t.Fatalf("enroll verify: %d %s", verify.Code, verify.Body.String())
	}
	login := postJSON(handler, "/api/admin/session/login", map[string]any{"code": totpCode(startResp.Secret, time.Now().UTC())})
	if login.Code != http.StatusOK {
		t.Fatalf("login: %d %s", login.Code, login.Body.String())
	}
	cookies := login.Result().Cookies()
	if len(cookies) == 0 {
		t.Fatalf("expected session cookie")
	}
	return cookies[0]
}

func createSubmittedSecureInputEvidence(t *testing.T, handler http.Handler, cookie *http.Cookie, targetName string) string {
	t.Helper()
	createReq := httptest.NewRequest(http.MethodPost, "/api/admin/approval-challenges/secure-input", strings.NewReader(`{
		"title":"Approval Evidence Input",
		"prompt":"Provide approval evidence.",
		"input_label":"Evidence",
		"submit_label":"Store Evidence",
		"target_name":"`+targetName+`"
	}`))
	createReq.Header.Set("Content-Type", "application/json")
	createReq.AddCookie(cookie)
	createW := httptest.NewRecorder()
	handler.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusCreated {
		t.Fatalf("create secure input evidence: %d %s", createW.Code, createW.Body.String())
	}
	var createResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode secure input evidence: %v", err)
	}
	submit := postForm(handler, "/approve/t/"+createResp.Token, map[string]string{
		"token": createResp.Token,
		"value": "operator-confirmed",
	})
	if submit.Code != http.StatusOK {
		t.Fatalf("submit secure input evidence: %d %s", submit.Code, submit.Body.String())
	}
	return createResp.Token
}
