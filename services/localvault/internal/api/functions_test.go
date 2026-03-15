package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"veilkey-localvault/internal/db"
)

func mustPerformJSON(t *testing.T, handler http.Handler, method, path string, body string) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}

func containsAll(haystack string, needles ...string) bool {
	for _, needle := range needles {
		if !strings.Contains(haystack, needle) {
			return false
		}
	}
	return true
}

func TestFunctionCRUDEndpoints(t *testing.T) {
	server := setupReencryptTestServer(t)
	handler := server.SetupRoutes()

	save := mustPerformJSON(t, handler, http.MethodPost, "/api/functions", `{
		"name":"gitlab/current-user",
		"scope":"LOCAL",
		"vault_hash":"56093730",
		"function_hash":"fn123456",
		"category":"gitlab",
		"command":"curl -sS https://example.test/{%{TOKEN}%}",
		"vars_json":"{\"TOKEN\":{\"ref\":\"VK:EXTERNAL:abcd1234\",\"class\":\"EXTERNAL\"}}"
	}`)
	if save.Code != http.StatusOK {
		t.Fatalf("save failed: %d %s", save.Code, save.Body.String())
	}

	list := mustPerformJSON(t, handler, http.MethodGet, "/api/functions", "")
	if list.Code != http.StatusOK {
		t.Fatalf("list failed: %d %s", list.Code, list.Body.String())
	}
	if body := list.Body.String(); !containsAll(body, `"count":1`, `"name":"gitlab/current-user"`) {
		t.Fatalf("unexpected list body: %s", body)
	}

	get := mustPerformJSON(t, handler, http.MethodGet, "/api/functions/gitlab/current-user", "")
	if get.Code != http.StatusOK {
		t.Fatalf("get failed: %d %s", get.Code, get.Body.String())
	}
	if body := get.Body.String(); !containsAll(body, `"function_hash":"fn123456"`, `"vault_hash":"56093730"`) {
		t.Fatalf("unexpected get body: %s", body)
	}

	del := mustPerformJSON(t, handler, http.MethodDelete, "/api/functions/gitlab/current-user", "")
	if del.Code != http.StatusOK {
		t.Fatalf("delete failed: %d %s", del.Code, del.Body.String())
	}

	missing := mustPerformJSON(t, handler, http.MethodGet, "/api/functions/gitlab/current-user", "")
	if missing.Code != http.StatusNotFound {
		t.Fatalf("expected missing after delete, got %d", missing.Code)
	}
}

func TestFunctionEndpointsReflectDatabaseRows(t *testing.T) {
	server := setupReencryptTestServer(t)
	if err := server.db.SaveFunction(&db.Function{
		Name:         "ops/healthcheck",
		Scope:        "TEST",
		VaultHash:    "56093730",
		FunctionHash: "fn999999",
		Category:     "ops",
		Command:      `curl -sS https://example.test/{%{TOKEN}%}`,
		VarsJSON:     `{"TOKEN":{"ref":"VK:LOCAL:abcd1234","class":"LOCAL"}}`,
	}); err != nil {
		t.Fatalf("SaveFunction: %v", err)
	}

	w := mustPerformJSON(t, server.SetupRoutes(), http.MethodGet, "/api/functions/ops/healthcheck", "")
	if w.Code != http.StatusOK {
		t.Fatalf("expected function row from db, got %d %s", w.Code, w.Body.String())
	}
	if body := w.Body.String(); !containsAll(body, `"name":"ops/healthcheck"`, `"scope":"TEST"`) {
		t.Fatalf("unexpected body: %s", body)
	}
}

func TestFunctionEndpointsRejectManualGlobalMutation(t *testing.T) {
	server := setupReencryptTestServer(t)
	handler := server.SetupRoutes()

	save := mustPerformJSON(t, handler, http.MethodPost, "/api/functions", `{
		"name":"gitlab/current-user",
		"scope":"GLOBAL",
		"vault_hash":"56093730",
		"function_hash":"fn-global-guard",
		"category":"gitlab",
		"command":"curl -sS https://example.test/{%{TOKEN}%}",
		"vars_json":"{\"TOKEN\":{\"ref\":\"VK:EXTERNAL:abcd1234\",\"class\":\"EXTERNAL\"}}"
	}`)
	if save.Code != http.StatusBadRequest {
		t.Fatalf("expected GLOBAL save rejection, got %d %s", save.Code, save.Body.String())
	}
	if !strings.Contains(save.Body.String(), "KeyCenter sync") {
		t.Fatalf("unexpected save rejection body: %s", save.Body.String())
	}

	if err := server.db.SaveFunction(&db.Function{
		Name:         "gitlab/current-user",
		Scope:        "GLOBAL",
		VaultHash:    "56093730",
		FunctionHash: "fn-global-materialized",
		Category:     "gitlab",
		Command:      `curl -sS https://example.test/{%{TOKEN}%}`,
		VarsJSON:     `{"TOKEN":{"ref":"VK:EXTERNAL:abcd1234","class":"EXTERNAL"}}`,
		Provenance:   "keycenter-sync",
	}); err != nil {
		t.Fatalf("SaveFunction global materialized: %v", err)
	}

	del := mustPerformJSON(t, handler, http.MethodDelete, "/api/functions/gitlab/current-user", "")
	if del.Code != http.StatusBadRequest {
		t.Fatalf("expected GLOBAL delete rejection, got %d %s", del.Code, del.Body.String())
	}
	if !strings.Contains(del.Body.String(), "KeyCenter sync") {
		t.Fatalf("unexpected delete rejection body: %s", del.Body.String())
	}
}

func TestFunctionListSupportsScopeFilter(t *testing.T) {
	server := setupReencryptTestServer(t)
	handler := server.SetupRoutes()

	for _, fn := range []db.Function{
		{
			Name:         "gitlab/current-user",
			Scope:        "LOCAL",
			VaultHash:    "56093730",
			FunctionHash: "fn-local-list",
			Category:     "gitlab",
			Command:      `curl -sS https://example.test/{%{TOKEN}%}`,
			VarsJSON:     `{"TOKEN":{"ref":"VK:LOCAL:abcd1234","class":"LOCAL"}}`,
		},
		{
			Name:         "ops/healthcheck",
			Scope:        "TEST",
			VaultHash:    "56093730",
			FunctionHash: "fn-test-list",
			Category:     "ops",
			Command:      `curl -sS https://example.test/{%{TOKEN}%}`,
			VarsJSON:     `{"TOKEN":{"ref":"VK:EXTERNAL:abcd1234","class":"EXTERNAL"}}`,
		},
	} {
		if err := server.db.SaveFunction(&fn); err != nil {
			t.Fatalf("SaveFunction(%s): %v", fn.Name, err)
		}
	}

	filtered := mustPerformJSON(t, handler, http.MethodGet, "/api/functions?scope=TEST", "")
	if filtered.Code != http.StatusOK {
		t.Fatalf("filtered list failed: %d %s", filtered.Code, filtered.Body.String())
	}
	if body := filtered.Body.String(); !containsAll(body, `"count":1`, `"name":"ops/healthcheck"`) {
		t.Fatalf("unexpected filtered list body: %s", body)
	}
	if strings.Contains(filtered.Body.String(), "gitlab/current-user") {
		t.Fatalf("LOCAL function should not appear in TEST filtered list: %s", filtered.Body.String())
	}

	invalid := mustPerformJSON(t, handler, http.MethodGet, "/api/functions?scope=INVALID", "")
	if invalid.Code != http.StatusBadRequest {
		t.Fatalf("expected invalid scope rejection, got %d %s", invalid.Code, invalid.Body.String())
	}
}
