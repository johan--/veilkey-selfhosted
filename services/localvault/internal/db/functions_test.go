package db

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func newFunctionTestDB(t *testing.T) *DB {
	t.Helper()
	dbPath := filepath.Join(t.TempDir(), "veilkey.db")
	database, err := New(dbPath)
	if err != nil {
		t.Fatalf("db.New: %v", err)
	}
	t.Cleanup(func() {
		_ = database.Close()
	})
	return database
}

func TestSaveAndGetFunction(t *testing.T) {
	database := newFunctionTestDB(t)

	fn := &Function{
		Name:         "gitlab-project-get",
		Scope:        "GLOBAL",
		VaultHash:    "56093730",
		FunctionHash: "fnhash001",
		Category:     "gitlab",
		Command:      `curl -H "PRIVATE-TOKEN: {%{GITLAB_TOKEN}%}" https://gitlab.example/api/v4/projects/{%{PROJECT_ID}%}`,
		VarsJSON:     `{"GITLAB_TOKEN":{"ref":"VK:EXTERNAL:abcd1234","class":"EXTERNAL"},"PROJECT_ID":{"ref":"VE:LOCAL:project_id","class":"LOCAL"}}`,
	}
	if err := database.SaveFunction(fn); err != nil {
		t.Fatalf("SaveFunction: %v", err)
	}

	got, err := database.GetFunction("gitlab-project-get")
	if err != nil {
		t.Fatalf("GetFunction: %v", err)
	}
	if got.Scope != "GLOBAL" {
		t.Fatalf("scope = %q, want GLOBAL", got.Scope)
	}
	if got.VaultHash != "56093730" {
		t.Fatalf("vault_hash = %q", got.VaultHash)
	}
	if got.FunctionHash != "fnhash001" {
		t.Fatalf("function_hash = %q", got.FunctionHash)
	}
	if !strings.Contains(got.VarsJSON, `"class":"EXTERNAL"`) {
		t.Fatalf("vars_json missing EXTERNAL class: %s", got.VarsJSON)
	}

	logs, err := database.ListFunctionLogs()
	if err != nil {
		t.Fatalf("ListFunctionLogs: %v", err)
	}
	if len(logs) != 1 {
		t.Fatalf("logs len = %d, want 1", len(logs))
	}
	if logs[0].Action != "save" || logs[0].Status != "ok" {
		t.Fatalf("unexpected log: %#v", logs[0])
	}
}

func TestCleanupExpiredTestFunctions(t *testing.T) {
	database := newFunctionTestDB(t)

	testFn := &Function{
		Name:         "gitlab-test",
		Scope:        "TEST",
		VaultHash:    "56093730",
		FunctionHash: "fnhash-test",
		Category:     "gitlab",
		Command:      `curl https://example.invalid/{%{TOKEN}%}`,
		VarsJSON:     `{"TOKEN":{"ref":"VK:EXTERNAL:deadbeef","class":"EXTERNAL"}}`,
	}
	if err := database.SaveFunction(testFn); err != nil {
		t.Fatalf("SaveFunction(test): %v", err)
	}
	if _, err := database.conn.Exec(`UPDATE functions SET created_at = ? WHERE name = ?`, time.Now().Add(-2*time.Hour), testFn.Name); err != nil {
		t.Fatalf("age test function: %v", err)
	}

	localFn := &Function{
		Name:         "gitlab-live",
		Scope:        "LOCAL",
		VaultHash:    "56093730",
		FunctionHash: "fnhash-live",
		Category:     "gitlab",
		Command:      `curl https://example.invalid/{%{TOKEN}%}`,
		VarsJSON:     `{"TOKEN":{"ref":"VK:EXTERNAL:beadfeed","class":"EXTERNAL"}}`,
	}
	if err := database.SaveFunction(localFn); err != nil {
		t.Fatalf("SaveFunction(local): %v", err)
	}

	deleted, err := database.CleanupExpiredTestFunctions(time.Now())
	if err != nil {
		t.Fatalf("CleanupExpiredTestFunctions: %v", err)
	}
	if deleted != 1 {
		t.Fatalf("deleted = %d, want 1", deleted)
	}

	if _, err := database.GetFunction("gitlab-test"); err == nil {
		t.Fatal("expected expired TEST function to be deleted")
	}
	if _, err := database.GetFunction("gitlab-live"); err != nil {
		t.Fatalf("expected LOCAL function to remain: %v", err)
	}

	logs, err := database.ListFunctionLogs()
	if err != nil {
		t.Fatalf("ListFunctionLogs: %v", err)
	}
	if len(logs) != 3 {
		t.Fatalf("logs len = %d, want 3", len(logs))
	}
	last := logs[len(logs)-1]
	if last.Action != "cleanup" || last.Status != "deleted" {
		t.Fatalf("unexpected cleanup log: %#v", last)
	}
}

func TestFunctionMetadataRoundTrip(t *testing.T) {
	database := newFunctionTestDB(t)

	testedAt := sql.NullTime{Time: time.Now().Add(-30 * time.Minute).UTC().Truncate(time.Second), Valid: true}
	runAt := sql.NullTime{Time: time.Now().Add(-10 * time.Minute).UTC().Truncate(time.Second), Valid: true}
	fn := &Function{
		Name:         "gitlab-project-list",
		Scope:        "VAULT",
		VaultHash:    "56093730",
		FunctionHash: "fnhash-meta",
		Category:     "gitlab",
		Command:      `curl https://gitlab.example/api/v4/projects`,
		VarsJSON:     `{"GITLAB_TOKEN":{"ref":"VK:LOCAL:gitlab_token","class":"key"}}`,
		Description:  "List projects from bound GitLab",
		TagsJSON:     `["gitlab","inventory"]`,
		Provenance:   "catalog",
		LastTestedAt: testedAt,
		LastRunAt:    runAt,
	}
	if err := database.SaveFunction(fn); err != nil {
		t.Fatalf("SaveFunction: %v", err)
	}

	got, err := database.GetFunction(fn.Name)
	if err != nil {
		t.Fatalf("GetFunction: %v", err)
	}
	if got.Description != fn.Description {
		t.Fatalf("Description = %q, want %q", got.Description, fn.Description)
	}
	if got.TagsJSON != fn.TagsJSON {
		t.Fatalf("TagsJSON = %q, want %q", got.TagsJSON, fn.TagsJSON)
	}
	if got.Provenance != fn.Provenance {
		t.Fatalf("Provenance = %q, want %q", got.Provenance, fn.Provenance)
	}
	if !got.LastTestedAt.Valid || !got.LastRunAt.Valid {
		t.Fatalf("expected last_tested_at and last_run_at to survive round trip: %#v", got)
	}
}

func TestSaveFunctionRejectsInvalidScope(t *testing.T) {
	database := newFunctionTestDB(t)

	err := database.SaveFunction(&Function{
		Name:         "bad-scope",
		Scope:        "INVALID",
		VaultHash:    "56093730",
		FunctionHash: "fn-invalid",
		Category:     "ops",
		Command:      `curl https://example.invalid/{%{TOKEN}%}`,
		VarsJSON:     `{"TOKEN":{"ref":"VK:EXTERNAL:deadbeef","class":"EXTERNAL"}}`,
	})
	if err == nil {
		t.Fatal("expected invalid scope error")
	}
	if !strings.Contains(err.Error(), "invalid function scope") {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestListFunctionsByScope(t *testing.T) {
	database := newFunctionTestDB(t)
	for _, fn := range []Function{
		{
			Name:         "ops/global",
			Scope:        "GLOBAL",
			VaultHash:    "56093730",
			FunctionHash: "fn-global-list",
			Category:     "ops",
			Command:      `curl https://example.invalid/{%{TOKEN}%}`,
			VarsJSON:     `{"TOKEN":{"ref":"VK:EXTERNAL:global","class":"EXTERNAL"}}`,
		},
		{
			Name:         "ops/test",
			Scope:        "TEST",
			VaultHash:    "56093730",
			FunctionHash: "fn-test-list",
			Category:     "ops",
			Command:      `curl https://example.invalid/{%{TOKEN}%}`,
			VarsJSON:     `{"TOKEN":{"ref":"VK:EXTERNAL:test","class":"EXTERNAL"}}`,
		},
	} {
		if err := database.SaveFunction(&fn); err != nil {
			t.Fatalf("SaveFunction(%s): %v", fn.Name, err)
		}
	}

	rows, err := database.ListFunctionsByScope("TEST")
	if err != nil {
		t.Fatalf("ListFunctionsByScope: %v", err)
	}
	if len(rows) != 1 || rows[0].Name != "ops/test" {
		t.Fatalf("unexpected TEST rows: %#v", rows)
	}
}
