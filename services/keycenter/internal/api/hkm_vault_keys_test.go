package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"veilkey-keycenter/internal/db"
)

func TestVaultGetCanonicalDetail(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-detail-agent", map[string]string{"APP_ENV": "prod"}, map[string]string{"API_KEY": "TEMP:deadbeef"})

	get := getJSON(handler, "/api/vaults/"+agentHash)
	if get.Code != http.StatusOK {
		t.Fatalf("get vault: expected 200, got %d: %s", get.Code, get.Body.String())
	}
	var resp struct {
		VaultRuntimeHash string `json:"vault_runtime_hash"`
		Label            string `json:"label"`
		VaultName        string `json:"vault_name"`
		SecretsCount     int    `json:"secrets_count"`
		ConfigsCount     int    `json:"configs_count"`
		Status           string `json:"status"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode vault response: %v", err)
	}
	if resp.VaultRuntimeHash != agentHash || resp.Label != "vault-detail-agent" || resp.VaultName != "vault-detail-agent" {
		t.Fatalf("unexpected vault response: %+v", resp)
	}
	if resp.SecretsCount != 1 || resp.ConfigsCount != 1 || resp.Status != "ok" {
		t.Fatalf("unexpected vault counters/status: %+v", resp)
	}
}

func TestVaultPatchUpdatesInventoryMetadata(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-meta-agent", map[string]string{}, nil)

	patch := patchJSON(handler, "/api/vaults/"+agentHash, map[string]any{
		"display_name": "Primary Vault",
		"description":  "main production vault",
		"tags":         []string{"prod", "critical"},
	})
	if patch.Code != http.StatusOK {
		t.Fatalf("patch vault: expected 200, got %d: %s", patch.Code, patch.Body.String())
	}

	get := getJSON(handler, "/api/vaults/"+agentHash)
	if get.Code != http.StatusOK {
		t.Fatalf("get vault: expected 200, got %d: %s", get.Code, get.Body.String())
	}
	var resp struct {
		DisplayName string   `json:"display_name"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode vault response: %v", err)
	}
	if resp.DisplayName != "Primary Vault" || resp.Description != "main production vault" || len(resp.Tags) != 2 {
		t.Fatalf("unexpected patched vault response: %+v", resp)
	}
}

func TestVaultListCanonicalRoute(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, firstHash := registerMockAgent(t, srv, "vault-list-alpha", map[string]string{}, nil)
	_, secondHash := registerMockAgent(t, srv, "vault-list-beta", map[string]string{}, nil)

	list := getJSON(handler, "/api/vaults")
	if list.Code != http.StatusOK {
		t.Fatalf("list vaults: expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		Vaults []map[string]interface{} `json:"vaults"`
		Count  int                      `json:"count"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode vault list response: %v", err)
	}
	if listResp.Count < 2 || len(listResp.Vaults) < 2 {
		t.Fatalf("expected at least 2 vaults, got %+v", listResp)
	}

	search := getJSON(handler, "/api/vaults?q=beta")
	if search.Code != http.StatusOK {
		t.Fatalf("search vaults: expected 200, got %d: %s", search.Code, search.Body.String())
	}
	var searchResp struct {
		Vaults []map[string]interface{} `json:"vaults"`
		Count  int                      `json:"count"`
	}
	if err := json.Unmarshal(search.Body.Bytes(), &searchResp); err != nil {
		t.Fatalf("decode vault search response: %v", err)
	}
	if searchResp.Count != 1 {
		t.Fatalf("expected 1 searched vault, got %+v", searchResp)
	}
	gotHash, _ := searchResp.Vaults[0]["vault_runtime_hash"].(string)
	if gotHash != secondHash && gotHash != firstHash {
		t.Fatalf("unexpected searched vault hash: %+v", searchResp.Vaults[0])
	}
	label, _ := searchResp.Vaults[0]["label"].(string)
	if label != "vault-list-beta" {
		t.Fatalf("unexpected searched vault label: %+v", searchResp.Vaults[0])
	}
}

func TestVaultListExcludesHostOnlyAgents(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-list-agent", map[string]string{}, nil)
	_, hostOnlyHash := registerMockAgent(t, srv, "vault-list-host-only", map[string]string{}, nil)
	if err := srv.db.UpdateAgentCapabilities("node-vault-list-host-only", "host-only", nil, nil); err != nil {
		t.Fatalf("UpdateAgentCapabilities(host-only): %v", err)
	}

	list := getJSON(handler, "/api/vaults")
	if list.Code != http.StatusOK {
		t.Fatalf("list vaults: expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		Vaults []map[string]interface{} `json:"vaults"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode vault list response: %v", err)
	}
	for _, item := range listResp.Vaults {
		gotHash, _ := item["vault_runtime_hash"].(string)
		if gotHash == hostOnlyHash {
			t.Fatalf("host-only vault should be excluded from /api/vaults: %+v", item)
		}
	}

	get := getJSON(handler, "/api/vaults/"+hostOnlyHash)
	if get.Code != http.StatusNotFound {
		t.Fatalf("expected host-only /api/vaults/{vault} to return 404, got %d: %s", get.Code, get.Body.String())
	}

	visible := false
	for _, item := range listResp.Vaults {
		gotHash, _ := item["vault_runtime_hash"].(string)
		if gotHash == agentHash {
			visible = true
			break
		}
	}
	if !visible {
		t.Fatalf("normal agent vault %s should stay visible", agentHash)
	}
}

func TestHeartbeatPreservesExistingHostOnlyCapabilitiesWhenOmitted(t *testing.T) {
	srv, _ := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-host-role", map[string]string{}, nil)
	agent, err := srv.db.GetAgentByHash(agentHash)
	if err != nil {
		t.Fatalf("GetAgentByHash: %v", err)
	}
	if err := srv.db.UpdateAgentCapabilities(agent.NodeID, "host-only", nil, nil); err != nil {
		t.Fatalf("UpdateAgentCapabilities(host-only): %v", err)
	}

	w := postJSON(srv.SetupRoutes(), "/api/agents/heartbeat", map[string]any{
		"vault_node_uuid": agent.NodeID,
		"label":           agent.Label,
		"vault_hash":      agent.VaultHash,
		"vault_name":      agent.VaultName,
		"ip":              agent.IP,
		"port":            agent.Port,
		"secrets_count":   agent.SecretsCount,
		"configs_count":   agent.ConfigsCount,
		"version":         agent.Version,
		"key_version":     agent.KeyVersion,
	})
	if w.Code != http.StatusOK {
		t.Fatalf("heartbeat: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	updated, err := srv.db.GetAgentByNodeID(agent.NodeID)
	if err != nil {
		t.Fatalf("GetAgentByNodeID: %v", err)
	}
	if !updated.HostEnabled || updated.LocalEnabled {
		t.Fatalf("capabilities = host:%v local:%v, want host-only", updated.HostEnabled, updated.LocalEnabled)
	}
}

func TestVaultListIncludesDualCapabilityAgents(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, dualHash := registerMockAgent(t, srv, "vault-dual-capability", map[string]string{}, nil)
	if err := srv.db.UpdateAgentCapabilities("node-vault-dual-capability", "host+local", nil, nil); err != nil {
		t.Fatalf("UpdateAgentCapabilities(host+local): %v", err)
	}

	list := getJSON(handler, "/api/vaults")
	if list.Code != http.StatusOK {
		t.Fatalf("list vaults: expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		Vaults []map[string]interface{} `json:"vaults"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode vault list response: %v", err)
	}
	found := false
	for _, item := range listResp.Vaults {
		gotHash, _ := item["vault_runtime_hash"].(string)
		if gotHash == dualHash {
			found = true
			hostEnabled, _ := item["host_enabled"].(bool)
			localEnabled, _ := item["local_enabled"].(bool)
			if !hostEnabled || !localEnabled {
				t.Fatalf("dual capability flags missing: %+v", item)
			}
		}
	}
	if !found {
		t.Fatalf("dual capability vault %s should remain visible in local vault list", dualHash)
	}
}

func TestVaultKeysCanonicalCRUD(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-keys-agent", map[string]string{}, nil)

	save := postJSON(handler, "/api/vaults/"+agentHash+"/keys", map[string]string{
		"name":  "API_KEY",
		"value": "super-secret-123",
	})
	if save.Code != 200 {
		t.Fatalf("save key: expected 200, got %d: %s", save.Code, save.Body.String())
	}
	var saveResp struct {
		Name  string `json:"name"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(save.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("decode save response: %v", err)
	}
	if saveResp.Name != "API_KEY" || saveResp.Token == "" {
		t.Fatalf("unexpected save response: %+v", saveResp)
	}

	list := getJSON(handler, "/api/vaults/"+agentHash+"/keys")
	if list.Code != 200 {
		t.Fatalf("list keys: expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		VaultRuntimeHash string                   `json:"vault_runtime_hash"`
		Secrets          []map[string]interface{} `json:"secrets"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.VaultRuntimeHash != agentHash || len(listResp.Secrets) != 1 {
		t.Fatalf("unexpected list response: %+v", listResp)
	}

	get := getJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY")
	if get.Code != 200 {
		t.Fatalf("get key: expected 200, got %d: %s", get.Code, get.Body.String())
	}
	var getResp struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getResp.Name != "API_KEY" || getResp.Value != "super-secret-123" {
		t.Fatalf("unexpected get response: %+v", getResp)
	}

	update := putJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY", map[string]string{
		"value": "super-secret-456",
	})
	if update.Code != 200 {
		t.Fatalf("update key: expected 200, got %d: %s", update.Code, update.Body.String())
	}

	getUpdated := getJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY")
	if getUpdated.Code != 200 {
		t.Fatalf("get updated key: expected 200, got %d: %s", getUpdated.Code, getUpdated.Body.String())
	}
	var getUpdatedResp struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(getUpdated.Body.Bytes(), &getUpdatedResp); err != nil {
		t.Fatalf("decode updated get response: %v", err)
	}
	if getUpdatedResp.Name != "API_KEY" || getUpdatedResp.Value != "super-secret-456" {
		t.Fatalf("unexpected updated get response: %+v", getUpdatedResp)
	}

	del := deleteJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY")
	if del.Code != 200 {
		t.Fatalf("delete key: expected 200, got %d: %s", del.Code, del.Body.String())
	}
}

func TestVaultKeyUpdateRejectsMismatchedBodyName(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-keys-update-agent", map[string]string{}, nil)
	update := putJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY", map[string]string{
		"name":  "OTHER_KEY",
		"value": "nope",
	})
	if update.Code != 400 {
		t.Fatalf("update mismatch: expected 400, got %d: %s", update.Code, update.Body.String())
	}
}

func TestVaultKeyFieldCanonicalCRUD(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-key-field-agent", nil, map[string]string{"GITHUB_KEY": "EXTERNAL:deadbeef"})

	saveKey := postJSON(handler, "/api/vaults/"+agentHash+"/keys", map[string]string{
		"name":  "GITHUB_KEY",
		"value": "ghp_example_token",
	})
	if saveKey.Code != 200 {
		t.Fatalf("save key: expected 200, got %d: %s", saveKey.Code, saveKey.Body.String())
	}

	putField := putJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/fields/OTP", map[string]string{
		"type":  "otp",
		"value": "123456",
	})
	if putField.Code != 200 {
		t.Fatalf("put field: expected 200, got %d: %s", putField.Code, putField.Body.String())
	}

	getField := getJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/fields/OTP")
	if getField.Code != 200 {
		t.Fatalf("get field: expected 200, got %d: %s", getField.Code, getField.Body.String())
	}
	var getFieldResp struct {
		Field string `json:"field"`
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(getField.Body.Bytes(), &getFieldResp); err != nil {
		t.Fatalf("decode field response: %v", err)
	}
	if getFieldResp.Field != "OTP" || getFieldResp.Type != "otp" || getFieldResp.Value != "123456" {
		t.Fatalf("unexpected field response: %+v", getFieldResp)
	}

	delField := deleteJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/fields/OTP")
	if delField.Code != 200 {
		t.Fatalf("delete field: expected 200, got %d: %s", delField.Code, delField.Body.String())
	}
}

func TestVaultKeyMetaAndFieldsCanonicalRoutes(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-key-meta-agent", nil, map[string]string{"GITHUB_KEY": "EXTERNAL:deadbeef"})

	putFields := putJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/fields", map[string]any{
		"fields": []map[string]string{
			{"key": "OTP", "type": "otp", "value": "123456"},
			{"key": "LOGIN_ID", "type": "login", "value": "octocat"},
		},
	})
	if putFields.Code != http.StatusOK {
		t.Fatalf("bulk put fields: expected 200, got %d: %s", putFields.Code, putFields.Body.String())
	}

	meta := getJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/meta")
	if meta.Code != http.StatusOK {
		t.Fatalf("get key meta: expected 200, got %d: %s", meta.Code, meta.Body.String())
	}
	var metaResp struct {
		Name        string                   `json:"name"`
		Token       string                   `json:"token"`
		Scope       string                   `json:"scope"`
		Status      string                   `json:"status"`
		FieldsCount int                      `json:"fields_count"`
		Fields      []map[string]interface{} `json:"fields"`
	}
	if err := json.Unmarshal(meta.Body.Bytes(), &metaResp); err != nil {
		t.Fatalf("decode meta response: %v", err)
	}
	if metaResp.Name != "GITHUB_KEY" || metaResp.Scope != "EXTERNAL" || metaResp.Status != "active" || metaResp.FieldsCount != 2 {
		t.Fatalf("unexpected meta response: %+v", metaResp)
	}
	if metaResp.Token == "" {
		t.Fatalf("expected canonical token in meta response: %+v", metaResp)
	}

	fields := getJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/fields")
	if fields.Code != http.StatusOK {
		t.Fatalf("get key fields: expected 200, got %d: %s", fields.Code, fields.Body.String())
	}
	var fieldsResp struct {
		Name        string                   `json:"name"`
		FieldsCount int                      `json:"fields_count"`
		Fields      []map[string]interface{} `json:"fields"`
	}
	if err := json.Unmarshal(fields.Body.Bytes(), &fieldsResp); err != nil {
		t.Fatalf("decode fields response: %v", err)
	}
	if fieldsResp.Name != "GITHUB_KEY" || fieldsResp.FieldsCount != 2 || len(fieldsResp.Fields) != 2 {
		t.Fatalf("unexpected fields response: %+v", fieldsResp)
	}
}

func TestVaultKeyMetaPatchAndSummaryRoutes(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-key-summary-agent", nil, map[string]string{"GITHUB_KEY": "EXTERNAL:deadbeef"})

	putFields := putJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/fields", map[string]any{
		"fields": []map[string]string{
			{"key": "OTP", "type": "otp", "value": "123456"},
		},
	})
	if putFields.Code != http.StatusOK {
		t.Fatalf("bulk put fields: expected 200, got %d: %s", putFields.Code, putFields.Body.String())
	}

	if err := srv.db.SaveBinding(&db.Binding{
		BindingID:    "bind-summary-1",
		BindingType:  "function",
		TargetName:   "gitlab/current-user",
		VaultHash:    "hash-vault-key-summary-agent",
		SecretName:   "GITHUB_KEY",
		RefCanonical: "VK:EXTERNAL:deadbeef",
		Required:     true,
	}); err != nil {
		t.Fatalf("SaveBinding: %v", err)
	}
	srv.saveAuditEvent("secret", "VK:EXTERNAL:deadbeef", "resolve", "api", agentHash, "", "resolve", nil, map[string]any{"ref": "VK:EXTERNAL:deadbeef"})

	patch := patchJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/meta", map[string]any{
		"display_name": "GitHub Token",
		"description":  "used by gitlab sync",
		"tags":         []string{"github", "external"},
	})
	if patch.Code != http.StatusOK {
		t.Fatalf("patch key meta: expected 200, got %d: %s", patch.Code, patch.Body.String())
	}

	meta := getJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/meta")
	if meta.Code != http.StatusOK {
		t.Fatalf("get key meta: expected 200, got %d: %s", meta.Code, meta.Body.String())
	}
	var metaResp struct {
		DisplayName string   `json:"display_name"`
		Description string   `json:"description"`
		Tags        []string `json:"tags"`
		UsageCount  int      `json:"usage_count"`
	}
	if err := json.Unmarshal(meta.Body.Bytes(), &metaResp); err != nil {
		t.Fatalf("decode meta response: %v", err)
	}
	if metaResp.DisplayName != "GitHub Token" || metaResp.Description != "used by gitlab sync" || len(metaResp.Tags) != 2 || metaResp.UsageCount != 1 {
		t.Fatalf("unexpected meta payload: %+v", metaResp)
	}

	summary := getJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/summary")
	if summary.Code != http.StatusOK {
		t.Fatalf("get key summary: expected 200, got %d: %s", summary.Code, summary.Body.String())
	}
	var summaryResp struct {
		Vault struct {
			VaultRuntimeHash string `json:"vault_runtime_hash"`
		} `json:"vault"`
		Key struct {
			DisplayName string   `json:"display_name"`
			Description string   `json:"description"`
			Tags        []string `json:"tags"`
		} `json:"key"`
		BindingsCount    int `json:"bindings_count"`
		BindingsTotal    int `json:"bindings_total"`
		UsageCount       int `json:"usage_count"`
		RecentAuditCount int `json:"recent_audit_count"`
	}
	if err := json.Unmarshal(summary.Body.Bytes(), &summaryResp); err != nil {
		t.Fatalf("decode summary response: %v", err)
	}
	if summaryResp.Vault.VaultRuntimeHash != agentHash ||
		summaryResp.Key.DisplayName != "GitHub Token" ||
		summaryResp.Key.Description != "used by gitlab sync" ||
		len(summaryResp.Key.Tags) != 2 ||
		summaryResp.BindingsCount != 1 ||
		summaryResp.BindingsTotal != 1 ||
		summaryResp.UsageCount != 1 ||
		summaryResp.RecentAuditCount == 0 {
		t.Fatalf("unexpected summary payload: %+v", summaryResp)
	}
}

func TestVaultKeyFieldUpdateRejectsMismatchedBodyKey(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-key-field-mismatch-agent", nil, map[string]string{"GITHUB_KEY": "EXTERNAL:deadbeef"})

	saveKey := postJSON(handler, "/api/vaults/"+agentHash+"/keys", map[string]string{
		"name":  "GITHUB_KEY",
		"value": "ghp_example_token",
	})
	if saveKey.Code != 200 {
		t.Fatalf("save key: expected 200, got %d: %s", saveKey.Code, saveKey.Body.String())
	}

	putField := putJSON(handler, "/api/vaults/"+agentHash+"/keys/GITHUB_KEY/fields/OTP", map[string]string{
		"key":   "LOGIN_ID",
		"type":  "otp",
		"value": "123456",
	})
	if putField.Code != 400 {
		t.Fatalf("put field mismatch: expected 400, got %d: %s", putField.Code, putField.Body.String())
	}
}

func TestVaultKeyActivateIsExplicitlyNotImplemented(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-key-activate-agent", map[string]string{}, nil)
	activate := postJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/activate", map[string]string{
		"scope": "LOCAL",
	})
	if activate.Code != http.StatusNotImplemented {
		t.Fatalf("activate expected 501, got %d: %s", activate.Code, activate.Body.String())
	}
}

func TestVaultKeyUsageAndAuditRoutes(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-key-usage-agent", map[string]string{}, nil)
	save := postJSON(handler, "/api/vaults/"+agentHash+"/keys", map[string]string{
		"name":  "API_KEY",
		"value": "super-secret-123",
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save key: expected 200, got %d: %s", save.Code, save.Body.String())
	}
	var saveResp struct {
		Ref   string `json:"ref"`
		Token string `json:"token"`
	}
	if err := json.Unmarshal(save.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("decode save response: %v", err)
	}
	if saveResp.Token == "" {
		t.Fatalf("expected canonical token, got %+v", saveResp)
	}

	if err := srv.db.SaveBinding(&db.Binding{
		BindingID:    "bind-1",
		BindingType:  "function",
		TargetName:   "gitlab/current-user",
		VaultHash:    "hash-vault-key-usage-agent",
		SecretName:   "API_KEY",
		FieldKey:     "",
		RefCanonical: saveResp.Token,
		Required:     true,
	}); err != nil {
		t.Fatalf("SaveBinding(bind-1): %v", err)
	}
	if err := srv.db.SaveBinding(&db.Binding{
		BindingID:    "bind-2",
		BindingType:  "workflow",
		TargetName:   "deploy/prod",
		VaultHash:    "hash-vault-key-usage-agent",
		SecretName:   "API_KEY",
		FieldKey:     "",
		RefCanonical: saveResp.Token,
		Required:     true,
	}); err != nil {
		t.Fatalf("SaveBinding(bind-2): %v", err)
	}
	srv.saveAuditEvent("secret", saveResp.Token, "resolve", "api", agentHash, "", "resolve", nil, map[string]any{
		"ref": saveResp.Token,
	})

	usage := getJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/usage")
	if usage.Code != http.StatusOK {
		t.Fatalf("usage expected 200, got %d: %s", usage.Code, usage.Body.String())
	}
	var usageResp struct {
		Token      string                   `json:"token"`
		UsageCount int64                    `json:"usage_count"`
		Bindings   []map[string]interface{} `json:"bindings"`
		Count      int                      `json:"count"`
	}
	if err := json.Unmarshal(usage.Body.Bytes(), &usageResp); err != nil {
		t.Fatalf("decode usage response: %v", err)
	}
	if usageResp.Token != saveResp.Token || usageResp.UsageCount != 2 || usageResp.Count != 2 || len(usageResp.Bindings) != 2 {
		t.Fatalf("unexpected usage response: %+v", usageResp)
	}

	bindings := getJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/bindings")
	if bindings.Code != http.StatusOK {
		t.Fatalf("bindings expected 200, got %d: %s", bindings.Code, bindings.Body.String())
	}
	var bindingsResp struct {
		Token    string                   `json:"token"`
		Bindings []map[string]interface{} `json:"bindings"`
		Count    int                      `json:"count"`
	}
	if err := json.Unmarshal(bindings.Body.Bytes(), &bindingsResp); err != nil {
		t.Fatalf("decode bindings response: %v", err)
	}
	if bindingsResp.Token != saveResp.Token || bindingsResp.Count != 2 || len(bindingsResp.Bindings) != 2 {
		t.Fatalf("unexpected bindings response: %+v", bindingsResp)
	}

	audit := getJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/audit")
	if audit.Code != http.StatusOK {
		t.Fatalf("key audit expected 200, got %d: %s", audit.Code, audit.Body.String())
	}
	var auditResp struct {
		Token  string                   `json:"token"`
		Events []map[string]interface{} `json:"events"`
		Count  int                      `json:"count"`
	}
	if err := json.Unmarshal(audit.Body.Bytes(), &auditResp); err != nil {
		t.Fatalf("decode key audit response: %v", err)
	}
	if auditResp.Token != saveResp.Token || auditResp.Count == 0 || len(auditResp.Events) == 0 {
		t.Fatalf("unexpected key audit response: %+v", auditResp)
	}
}

func TestVaultKeyBindingSaveAndDeleteRoutes(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-key-binding-agent", map[string]string{}, nil)
	save := postJSON(handler, "/api/vaults/"+agentHash+"/keys", map[string]string{
		"name":  "API_KEY",
		"value": "super-secret-123",
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save key: expected 200, got %d: %s", save.Code, save.Body.String())
	}

	createBinding := postJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/bindings", map[string]any{
		"binding_type": "function",
		"target_name":  "gitlab/current-user",
		"field_key":    "OTP",
		"required":     false,
	})
	if createBinding.Code != http.StatusOK {
		t.Fatalf("create binding: expected 200, got %d: %s", createBinding.Code, createBinding.Body.String())
	}
	var createResp struct {
		BindingID        string `json:"binding_id"`
		BindingType      string `json:"binding_type"`
		TargetName       string `json:"target_name"`
		FieldKey         string `json:"field_key"`
		Required         bool   `json:"required"`
		VaultRuntimeHash string `json:"vault_runtime_hash"`
	}
	if err := json.Unmarshal(createBinding.Body.Bytes(), &createResp); err != nil {
		t.Fatalf("decode create binding response: %v", err)
	}
	if createResp.BindingID == "" || createResp.BindingType != "function" || createResp.TargetName != "gitlab/current-user" || createResp.FieldKey != "OTP" || createResp.Required {
		t.Fatalf("unexpected create binding response: %+v", createResp)
	}

	bindings := getJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/bindings")
	if bindings.Code != http.StatusOK {
		t.Fatalf("bindings expected 200, got %d: %s", bindings.Code, bindings.Body.String())
	}
	var bindingsResp struct {
		Count    int                      `json:"count"`
		Bindings []map[string]interface{} `json:"bindings"`
	}
	if err := json.Unmarshal(bindings.Body.Bytes(), &bindingsResp); err != nil {
		t.Fatalf("decode bindings response: %v", err)
	}
	if bindingsResp.Count != 1 || len(bindingsResp.Bindings) != 1 {
		t.Fatalf("unexpected bindings after create: %+v", bindingsResp)
	}

	deleteBinding := deleteJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/bindings/"+createResp.BindingID)
	if deleteBinding.Code != http.StatusOK {
		t.Fatalf("delete binding: expected 200, got %d: %s", deleteBinding.Code, deleteBinding.Body.String())
	}

	bindingsAfterDelete := getJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/bindings")
	if bindingsAfterDelete.Code != http.StatusOK {
		t.Fatalf("bindings after delete expected 200, got %d: %s", bindingsAfterDelete.Code, bindingsAfterDelete.Body.String())
	}
	var bindingsAfterDeleteResp struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(bindingsAfterDelete.Body.Bytes(), &bindingsAfterDeleteResp); err != nil {
		t.Fatalf("decode bindings after delete response: %v", err)
	}
	if bindingsAfterDeleteResp.Count != 0 {
		t.Fatalf("expected 0 bindings after delete, got %+v", bindingsAfterDeleteResp)
	}
}

func TestVaultKeyBindingsReplaceAndDeleteAllRoutes(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-key-binding-bulk-agent", map[string]string{}, nil)
	save := postJSON(handler, "/api/vaults/"+agentHash+"/keys", map[string]string{
		"name":  "API_KEY",
		"value": "super-secret-123",
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save key: expected 200, got %d: %s", save.Code, save.Body.String())
	}

	replace := putJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/bindings", map[string]any{
		"bindings": []map[string]any{
			{"binding_type": "function", "target_name": "gitlab/current-user"},
			{"binding_type": "workflow", "target_name": "deploy/prod", "required": false},
		},
	})
	if replace.Code != http.StatusOK {
		t.Fatalf("replace bindings: expected 200, got %d: %s", replace.Code, replace.Body.String())
	}
	var replaceResp struct {
		Saved int `json:"saved"`
	}
	if err := json.Unmarshal(replace.Body.Bytes(), &replaceResp); err != nil {
		t.Fatalf("decode replace response: %v", err)
	}
	if replaceResp.Saved != 2 {
		t.Fatalf("unexpected replace response: %+v", replaceResp)
	}

	list := getJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/bindings")
	if list.Code != http.StatusOK {
		t.Fatalf("list bindings after replace: expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.Count != 2 {
		t.Fatalf("expected 2 bindings after replace, got %+v", listResp)
	}

	deleteAll := deleteJSON(handler, "/api/vaults/"+agentHash+"/keys/API_KEY/bindings")
	if deleteAll.Code != http.StatusOK {
		t.Fatalf("delete all bindings: expected 200, got %d: %s", deleteAll.Code, deleteAll.Body.String())
	}
	var deleteAllResp struct {
		Deleted int `json:"deleted"`
	}
	if err := json.Unmarshal(deleteAll.Body.Bytes(), &deleteAllResp); err != nil {
		t.Fatalf("decode delete all response: %v", err)
	}
	if deleteAllResp.Deleted != 2 {
		t.Fatalf("expected 2 deleted bindings, got %+v", deleteAllResp)
	}
}

func TestVaultAuditRoute(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "vault-audit-agent", map[string]string{}, nil)
	agent, err := srv.db.GetAgentByHash(agentHash)
	if err != nil {
		t.Fatalf("GetAgentByHash: %v", err)
	}
	srv.saveAuditEvent("vault", agent.NodeID, "schedule_rotation", "api", "tester", "", "test", nil, map[string]any{
		"vault_runtime_hash": agentHash,
	})

	w := getJSON(handler, "/api/vaults/"+agentHash+"/audit")
	if w.Code != http.StatusOK {
		t.Fatalf("vault audit expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		VaultRuntimeHash string                   `json:"vault_runtime_hash"`
		VaultNodeUUID    string                   `json:"vault_node_uuid"`
		Events           []map[string]interface{} `json:"events"`
		Count            int                      `json:"count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode vault audit response: %v", err)
	}
	if resp.VaultRuntimeHash != agentHash || resp.VaultNodeUUID != agent.NodeID || resp.Count == 0 || len(resp.Events) == 0 {
		t.Fatalf("unexpected vault audit response: %+v", resp)
	}
}
