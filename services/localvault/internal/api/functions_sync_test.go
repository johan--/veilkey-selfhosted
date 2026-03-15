package api

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"veilkey-localvault/internal/db"
)

func TestSyncGlobalFunctionsUpsertsAndPrunesGlobalCopies(t *testing.T) {
	server := setupReencryptTestServer(t)
	server.SetIdentity(&NodeIdentity{
		NodeID:    "56093730-0000-0000-0000-000000000001",
		VaultHash: "56093730",
		VaultName: "veilkey-localvault",
	})

	if err := server.db.SaveFunction(&db.Function{
		Name:         "gitlab/obsolete",
		Scope:        "GLOBAL",
		VaultHash:    "56093730",
		FunctionHash: "fn-obsolete",
		Category:     "gitlab",
		Command:      `curl -sS https://obsolete.test/{%{TOKEN}%}`,
		VarsJSON:     `{"TOKEN":{"ref":"VK:EXTERNAL:obsolete","class":"EXTERNAL"}}`,
	}); err != nil {
		t.Fatalf("SaveFunction obsolete: %v", err)
	}
	if err := server.db.SaveFunction(&db.Function{
		Name:         "ops/local-only",
		Scope:        "LOCAL",
		VaultHash:    "56093730",
		FunctionHash: "fn-local",
		Category:     "ops",
		Command:      `curl -sS https://local.test/{%{TOKEN}%}`,
		VarsJSON:     `{"TOKEN":{"ref":"VK:LOCAL:keepme","class":"LOCAL"}}`,
	}); err != nil {
		t.Fatalf("SaveFunction local: %v", err)
	}

	globalAPI := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"functions":[
				{
					"name":"gitlab/current-user",
					"scope":"GLOBAL",
					"vault_hash":"",
					"function_hash":"fn-global-1",
					"category":"gitlab",
					"command":"curl -sS https://gitlab.example/{%{TOKEN}%}",
					"vars_json":"{\"TOKEN\":{\"ref\":\"VK:EXTERNAL:abcd1234\",\"class\":\"EXTERNAL\"}}"
				}
			]
		}`))
	}))
	defer globalAPI.Close()

	upserted, deleted, err := server.SyncGlobalFunctions(globalAPI.URL)
	if err != nil {
		t.Fatalf("SyncGlobalFunctions: %v", err)
	}
	if upserted != 1 {
		t.Fatalf("expected 1 upserted global function, got %d", upserted)
	}
	if deleted != 1 {
		t.Fatalf("expected 1 deleted global function, got %d", deleted)
	}

	globalFn, err := server.db.GetFunction("gitlab/current-user")
	if err != nil {
		t.Fatalf("GetFunction synced global: %v", err)
	}
	if globalFn.Scope != "GLOBAL" {
		t.Fatalf("expected GLOBAL scope, got %s", globalFn.Scope)
	}
	if globalFn.VaultHash != "56093730" {
		t.Fatalf("expected materialized vault hash 56093730, got %s", globalFn.VaultHash)
	}

	if _, err := server.db.GetFunction("gitlab/obsolete"); err == nil {
		t.Fatal("expected obsolete global function to be pruned")
	}
	if _, err := server.db.GetFunction("ops/local-only"); err != nil {
		t.Fatalf("expected local-only function to remain: %v", err)
	}
}
