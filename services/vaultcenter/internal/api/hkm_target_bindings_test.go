package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"veilkey-vaultcenter/internal/db"
)

func TestTargetBindingsListReplaceAndImpactRoutes(t *testing.T) {
	srv, handler := setupHKMServer(t)

	first := targetSecretCatalogFixture()
	second := targetSecretCatalogFixture()
	second.SecretCanonicalID = "vh-a:VE:EXTERNAL:APP_URL"
	second.SecretName = "APP_URL"
	second.DisplayName = "App URL"
	second.Class = "config"
	second.Scope = "EXTERNAL"
	second.RefCanonical = "VE:EXTERNAL:APP_URL"
	if err := srv.db.SaveSecretCatalog(&first); err != nil {
		t.Fatalf("SaveSecretCatalog(first): %v", err)
	}
	if err := srv.db.SaveSecretCatalog(&second); err != nil {
		t.Fatalf("SaveSecretCatalog(second): %v", err)
	}

	replace := putJSON(handler, "/api/targets/function/gitlab%2Fcurrent-user/bindings", map[string]any{
		"bindings": []map[string]any{
			{"ref_canonical": "VK:LOCAL:deadbeef"},
			{"ref_canonical": "VE:EXTERNAL:APP_URL", "required": false},
		},
	})
	if replace.Code != http.StatusOK {
		t.Fatalf("replace target bindings: expected 200, got %d: %s", replace.Code, replace.Body.String())
	}

	list := getJSON(handler, "/api/targets/function/gitlab%2Fcurrent-user/bindings")
	if list.Code != http.StatusOK {
		t.Fatalf("list target bindings: expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		Count    int                      `json:"count"`
		Bindings []map[string]interface{} `json:"bindings"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if listResp.Count != 2 || len(listResp.Bindings) != 2 {
		t.Fatalf("unexpected list response: %+v", listResp)
	}

	impact := getJSON(handler, "/api/targets/function/gitlab%2Fcurrent-user/impact")
	if impact.Code != http.StatusOK {
		t.Fatalf("impact target bindings: expected 200, got %d: %s", impact.Code, impact.Body.String())
	}
	var impactResp struct {
		Count int                      `json:"count"`
		Refs  []map[string]interface{} `json:"refs"`
	}
	if err := json.Unmarshal(impact.Body.Bytes(), &impactResp); err != nil {
		t.Fatalf("decode impact response: %v", err)
	}
	if impactResp.Count != 2 || len(impactResp.Refs) != 2 {
		t.Fatalf("unexpected impact response: %+v", impactResp)
	}

	summary := getJSON(handler, "/api/targets/function/gitlab%2Fcurrent-user/summary")
	if summary.Code != http.StatusOK {
		t.Fatalf("summary target bindings: expected 200, got %d: %s", summary.Code, summary.Body.String())
	}
	var summaryResp struct {
		BindingsCount   int `json:"bindings_count"`
		BindingsTotal   int `json:"bindings_total"`
		UniqueRefsCount int `json:"unique_refs_count"`
		VaultsCount     int `json:"vaults_count"`
	}
	if err := json.Unmarshal(summary.Body.Bytes(), &summaryResp); err != nil {
		t.Fatalf("decode summary response: %v", err)
	}
	if summaryResp.BindingsCount != 2 || summaryResp.BindingsTotal != 2 || summaryResp.UniqueRefsCount != 2 || summaryResp.VaultsCount != 1 {
		t.Fatalf("unexpected summary response: %+v", summaryResp)
	}

	del := deleteJSON(handler, "/api/targets/function/gitlab%2Fcurrent-user/bindings")
	if del.Code != http.StatusOK {
		t.Fatalf("delete target bindings: expected 200, got %d: %s", del.Code, del.Body.String())
	}
	var delResp struct {
		Deleted int `json:"deleted"`
	}
	if err := json.Unmarshal(del.Body.Bytes(), &delResp); err != nil {
		t.Fatalf("decode delete response: %v", err)
	}
	if delResp.Deleted != 2 {
		t.Fatalf("unexpected delete response: %+v", delResp)
	}
}

func TestTargetBindingsReplaceRejectsUnknownRef(t *testing.T) {
	_, handler := setupHKMServer(t)

	replace := putJSON(handler, "/api/targets/function/gitlab%2Fcurrent-user/bindings", map[string]any{
		"bindings": []map[string]any{
			{"ref_canonical": "VK:LOCAL:missing"},
		},
	})
	if replace.Code != http.StatusNotFound {
		t.Fatalf("replace target bindings expected 404, got %d: %s", replace.Code, replace.Body.String())
	}
}

func TestTargetSummaryCountsMultipleVaultsAndMixedScopes(t *testing.T) {
	srv, handler := setupHKMServer(t)

	first := targetSecretCatalogFixture()
	second := targetSecretCatalogFixture()
	second.SecretCanonicalID = "vh-b:VK:TEMP:feedface"
	second.SecretName = "STAGING_TOKEN"
	second.DisplayName = "Staging Token"
	second.Scope = "TEMP"
	second.Status = "temp"
	second.VaultNodeUUID = "node-b"
	second.VaultRuntimeHash = "agent-b"
	second.VaultHash = "vh-b"
	second.RefCanonical = "VK:TEMP:feedface"
	if err := srv.db.SaveSecretCatalog(&first); err != nil {
		t.Fatalf("SaveSecretCatalog(first): %v", err)
	}
	if err := srv.db.SaveSecretCatalog(&second); err != nil {
		t.Fatalf("SaveSecretCatalog(second): %v", err)
	}

	replace := putJSON(handler, "/api/targets/workflow/deploy%2Fprod/bindings", map[string]any{
		"bindings": []map[string]any{
			{"ref_canonical": "VK:LOCAL:deadbeef"},
			{"ref_canonical": "VK:TEMP:feedface"},
		},
	})
	if replace.Code != http.StatusOK {
		t.Fatalf("replace target bindings: expected 200, got %d: %s", replace.Code, replace.Body.String())
	}

	summary := getJSON(handler, "/api/targets/workflow/deploy%2Fprod/summary")
	if summary.Code != http.StatusOK {
		t.Fatalf("summary target bindings: expected 200, got %d: %s", summary.Code, summary.Body.String())
	}
	var summaryResp struct {
		BindingsCount   int `json:"bindings_count"`
		UniqueRefsCount int `json:"unique_refs_count"`
		VaultsCount     int `json:"vaults_count"`
	}
	if err := json.Unmarshal(summary.Body.Bytes(), &summaryResp); err != nil {
		t.Fatalf("decode summary response: %v", err)
	}
	if summaryResp.BindingsCount != 2 || summaryResp.UniqueRefsCount != 2 || summaryResp.VaultsCount != 2 {
		t.Fatalf("unexpected multi-vault summary response: %+v", summaryResp)
	}
}

func targetSecretCatalogFixture() db.SecretCatalog {
	return db.SecretCatalog{
		SecretCanonicalID: "vh-a:VK:LOCAL:deadbeef",
		SecretName:        "OPENAI_API_KEY",
		DisplayName:       "OpenAI API Key",
		Description:       "primary provider credential",
		TagsJSON:          `["ai","prod"]`,
		Class:             "key",
		Scope:             "LOCAL",
		Status:            "active",
		VaultNodeUUID:     "node-a",
		VaultRuntimeHash:  "agent-a",
		VaultHash:         "vh-a",
		RefCanonical:      "VK:LOCAL:deadbeef",
		FieldsPresentJSON: `["OTP"]`,
		BindingCount:      1,
	}
}
