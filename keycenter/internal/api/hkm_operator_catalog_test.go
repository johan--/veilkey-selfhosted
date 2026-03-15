package api

import (
	"encoding/json"
	"testing"
	"veilkey-keycenter/internal/db"
)

func TestVaultInventoryEndpoint(t *testing.T) {
	srv, handler := setupHKMServer(t)

	if err := srv.db.UpsertAgent("node-a", "agent-a", "vh-a", "vault-a", "10.0.0.10", 10180, 1, 1, 1, 1); err != nil {
		t.Fatalf("UpsertAgent: %v", err)
	}
	if err := srv.db.UpdateAgentDEK("node-a", "agent-a", []byte("dek"), []byte("nonce")); err != nil {
		t.Fatalf("UpdateAgentDEK: %v", err)
	}

	w := getJSON(handler, "/api/vault-inventory")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Vaults []struct {
			VaultRuntimeHash string `json:"vault_runtime_hash"`
		} `json:"vaults"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Vaults) != 1 || resp.Vaults[0].VaultRuntimeHash != "agent-a" {
		t.Fatalf("unexpected vault inventory payload: %+v", resp)
	}
}

func TestVaultInventoryEndpointSupportsFilterAndPagination(t *testing.T) {
	srv, handler := setupHKMServer(t)

	if err := srv.db.UpsertAgent("node-a", "agent-a", "vh-a", "vault-a", "10.0.0.10", 10180, 1, 1, 1, 1); err != nil {
		t.Fatalf("UpsertAgent(a): %v", err)
	}
	if err := srv.db.UpdateAgentDEK("node-a", "agent-a", []byte("dek"), []byte("nonce")); err != nil {
		t.Fatalf("UpdateAgentDEK(a): %v", err)
	}
	if err := srv.db.UpsertAgent("node-b", "agent-b", "vh-b", "vault-b", "10.0.0.11", 10180, 1, 1, 1, 1); err != nil {
		t.Fatalf("UpsertAgent(b): %v", err)
	}
	if err := srv.db.UpdateAgentDEK("node-b", "agent-b", []byte("dek"), []byte("nonce")); err != nil {
		t.Fatalf("UpdateAgentDEK(b): %v", err)
	}

	w := getJSON(handler, "/api/vault-inventory?vault_hash=vh-a&limit=1&offset=0")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Count != 1 || resp.TotalCount != 1 {
		t.Fatalf("unexpected pagination counts: %+v", resp)
	}
}

func TestSecretCatalogEndpoints(t *testing.T) {
	srv, handler := setupHKMServer(t)

	entry := secretCatalogFixture()
	if err := srv.db.SaveSecretCatalog(&entry); err != nil {
		t.Fatalf("SaveSecretCatalog: %v", err)
	}

	list := getJSON(handler, "/api/catalog/secrets")
	if list.Code != 200 {
		t.Fatalf("list expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		Secrets []struct {
			BindingCount int `json:"binding_count"`
			UsageCount   int `json:"usage_count"`
		} `json:"secrets"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Secrets) != 1 || listResp.Secrets[0].UsageCount != listResp.Secrets[0].BindingCount {
		t.Fatalf("expected usage_count alias in list response, got %+v", listResp)
	}

	get := getJSON(handler, "/api/catalog/secrets/VK:LOCAL:deadbeef")
	if get.Code != 200 {
		t.Fatalf("get expected 200, got %d: %s", get.Code, get.Body.String())
	}
	var getResp struct {
		BindingCount int `json:"binding_count"`
		UsageCount   int `json:"usage_count"`
	}
	if err := json.Unmarshal(get.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("decode get response: %v", err)
	}
	if getResp.UsageCount != getResp.BindingCount {
		t.Fatalf("expected usage_count alias in get response, got %+v", getResp)
	}
}

func TestSecretCatalogListSupportsFilterAndPagination(t *testing.T) {
	srv, handler := setupHKMServer(t)

	first := secretCatalogFixture()
	second := secretCatalogFixture()
	second.SecretCanonicalID = "vh-a:VE:LOCAL:app_url"
	second.SecretName = "APP_URL"
	second.DisplayName = "App URL"
	second.Class = "config"
	second.RefCanonical = "VE:LOCAL:APP_URL"
	if err := srv.db.SaveSecretCatalog(&first); err != nil {
		t.Fatalf("SaveSecretCatalog(first): %v", err)
	}
	if err := srv.db.SaveSecretCatalog(&second); err != nil {
		t.Fatalf("SaveSecretCatalog(second): %v", err)
	}

	w := getJSON(handler, "/api/catalog/secrets?class=config&q=APP&limit=1")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Count != 1 || resp.TotalCount != 1 {
		t.Fatalf("unexpected pagination counts: %+v", resp)
	}
}

func TestSecretCatalogShowsOriginalNameForTempAgentSecret(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "catalog-name-agent", map[string]string{}, nil)

	save := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "GEMINI_API_KEY",
		"value": "super-secret-123",
	})
	if save.Code != 200 {
		t.Fatalf("save secret: expected 200, got %d: %s", save.Code, save.Body.String())
	}

	var saveResp struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(save.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("decode save response: %v", err)
	}
	if saveResp.Ref == "" {
		t.Fatal("save ref should not be empty")
	}

	list := getJSON(handler, "/api/catalog/secrets?class=key&q=GEMINI")
	if list.Code != 200 {
		t.Fatalf("catalog list expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		Secrets []struct {
			SecretName string `json:"secret_name"`
			Ref        string `json:"ref_canonical"`
		} `json:"secrets"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode catalog list: %v", err)
	}
	if len(listResp.Secrets) != 1 {
		t.Fatalf("expected 1 catalog row, got %d", len(listResp.Secrets))
	}
	if listResp.Secrets[0].SecretName != "GEMINI_API_KEY" {
		t.Fatalf("catalog list secret_name = %q", listResp.Secrets[0].SecretName)
	}

	get := getJSON(handler, "/api/catalog/secrets/VK:TEMP:"+saveResp.Ref)
	if get.Code != 200 {
		t.Fatalf("catalog get expected 200, got %d: %s", get.Code, get.Body.String())
	}
	var entry db.SecretCatalog
	if err := json.Unmarshal(get.Body.Bytes(), &entry); err != nil {
		t.Fatalf("decode catalog get: %v", err)
	}
	if entry.SecretName != "GEMINI_API_KEY" {
		t.Fatalf("catalog get secret_name = %q", entry.SecretName)
	}
	if entry.DisplayName != "GEMINI_API_KEY" {
		t.Fatalf("catalog get display_name = %q", entry.DisplayName)
	}
}

func TestSecretCatalogPreservesOriginalNameAfterAgentListAndGetSync(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "catalog-sync-agent", map[string]string{}, nil)

	save := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "GEMINI_API_KEY",
		"value": "super-secret-123",
	})
	if save.Code != 200 {
		t.Fatalf("save secret: expected 200, got %d: %s", save.Code, save.Body.String())
	}

	var saveResp struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(save.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("decode save response: %v", err)
	}
	if saveResp.Ref == "" {
		t.Fatal("save ref should not be empty")
	}

	list := getJSON(handler, "/api/agents/"+agentHash+"/secrets")
	if list.Code != 200 {
		t.Fatalf("agent list expected 200, got %d: %s", list.Code, list.Body.String())
	}

	get := getJSON(handler, "/api/agents/"+agentHash+"/secrets/GEMINI_API_KEY")
	if get.Code != 200 {
		t.Fatalf("agent get expected 200, got %d: %s", get.Code, get.Body.String())
	}

	ref, err := srv.db.GetRef("VK:TEMP:" + saveResp.Ref)
	if err != nil {
		t.Fatalf("tracked ref missing after sync: %v", err)
	}
	if ref.SecretName != "GEMINI_API_KEY" {
		t.Fatalf("tracked ref secret_name = %q", ref.SecretName)
	}

	entry, err := srv.db.GetSecretCatalogByRef("VK:TEMP:" + saveResp.Ref)
	if err != nil {
		t.Fatalf("catalog get: %v", err)
	}
	if entry.SecretName != "GEMINI_API_KEY" {
		t.Fatalf("catalog get secret_name = %q", entry.SecretName)
	}
	if entry.DisplayName != "GEMINI_API_KEY" {
		t.Fatalf("catalog get display_name = %q", entry.DisplayName)
	}
}

func TestBindingsEndpointRequiresQueryParams(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/catalog/bindings")
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBindingsEndpointListsRows(t *testing.T) {
	srv, handler := setupHKMServer(t)

	entry := bindingFixture("bind-1", "function", "gitlab/current-user", "VK:LOCAL:deadbeef")
	if err := srv.db.SaveBinding(&entry); err != nil {
		t.Fatalf("SaveBinding: %v", err)
	}

	w := getJSON(handler, "/api/catalog/bindings?binding_type=function&target_name=gitlab/current-user")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestBindingsEndpointListsRowsByRefCanonical(t *testing.T) {
	srv, handler := setupHKMServer(t)

	first := bindingFixture("bind-1", "function", "gitlab/current-user", "VK:LOCAL:deadbeef")
	second := bindingFixture("bind-2", "workflow", "deploy/prod", "VK:LOCAL:deadbeef")
	if err := srv.db.SaveBinding(&first); err != nil {
		t.Fatalf("SaveBinding(first): %v", err)
	}
	if err := srv.db.SaveBinding(&second); err != nil {
		t.Fatalf("SaveBinding(second): %v", err)
	}

	w := getJSON(handler, "/api/catalog/bindings?ref_canonical=VK:LOCAL:deadbeef")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Count != 2 || resp.TotalCount != 2 {
		t.Fatalf("unexpected ref-scoped counts: %+v", resp)
	}
}

func TestBindingsEndpointSupportsFilterAndPagination(t *testing.T) {
	srv, handler := setupHKMServer(t)

	first := bindingFixture("bind-1", "function", "gitlab/current-user", "VK:LOCAL:deadbeef")
	second := bindingFixture("bind-2", "function", "gitlab/current-user", "VE:LOCAL:APP_URL")
	second.SecretName = "APP_URL"
	second.VaultHash = "vh-b"
	if err := srv.db.SaveBinding(&first); err != nil {
		t.Fatalf("SaveBinding(first): %v", err)
	}
	if err := srv.db.SaveBinding(&second); err != nil {
		t.Fatalf("SaveBinding(second): %v", err)
	}

	w := getJSON(handler, "/api/catalog/bindings?binding_type=function&target_name=gitlab/current-user&vault_hash=vh-b&limit=1")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Count != 1 || resp.TotalCount != 1 {
		t.Fatalf("unexpected pagination counts: %+v", resp)
	}
}

func TestAuditEventsEndpointRequiresQueryParams(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/catalog/audit")
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuditEventsEndpointListsRows(t *testing.T) {
	srv, handler := setupHKMServer(t)

	srv.saveAuditEvent("secret", "VK:LOCAL:deadbeef", "resolve", "api", "agent-a", "", "resolve", nil, map[string]any{
		"ref": "VK:LOCAL:deadbeef",
	})

	w := getJSON(handler, "/api/catalog/audit?entity_type=secret&entity_id=VK:LOCAL:deadbeef")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestAuditEventsEndpointSupportsPagination(t *testing.T) {
	srv, handler := setupHKMServer(t)

	srv.saveAuditEvent("secret", "VK:LOCAL:deadbeef", "resolve", "api", "agent-a", "", "resolve", nil, map[string]any{"n": 1})
	srv.saveAuditEvent("secret", "VK:LOCAL:deadbeef", "resolve", "api", "agent-a", "", "resolve", nil, map[string]any{"n": 2})

	w := getJSON(handler, "/api/catalog/audit?entity_type=secret&entity_id=VK:LOCAL:deadbeef&limit=1")
	if w.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	var resp struct {
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Count != 1 || resp.TotalCount != 2 {
		t.Fatalf("unexpected pagination counts: %+v", resp)
	}
}

func TestCatalogEndpointsRejectInvalidPagination(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/catalog/secrets?limit=-1")
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
}

func secretCatalogFixture() db.SecretCatalog {
	return db.SecretCatalog{
		SecretCanonicalID: "vh-a:VK:LOCAL:deadbeef",
		SecretName:        "API_KEY",
		DisplayName:       "API Key",
		Class:             "key",
		Scope:             "LOCAL",
		Status:            "active",
		VaultNodeUUID:     "node-a",
		VaultRuntimeHash:  "agent-a",
		VaultHash:         "vh-a",
		RefCanonical:      "VK:LOCAL:deadbeef",
	}
}

func bindingFixture(bindingID, bindingType, targetName, refCanonical string) db.Binding {
	return db.Binding{
		BindingID:    bindingID,
		BindingType:  bindingType,
		TargetName:   targetName,
		VaultHash:    "vh-a",
		SecretName:   "API_KEY",
		RefCanonical: refCanonical,
		Required:     true,
	}
}
