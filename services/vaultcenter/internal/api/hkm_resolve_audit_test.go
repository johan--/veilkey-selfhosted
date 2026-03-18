package api

import (
	"encoding/json"
	"testing"
)

func TestResolveTrackedRefUpdatesCatalogRevealAudit(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "resolve-audit-agent", map[string]string{}, nil)

	save := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "OPENAI_API_KEY",
		"value": "super-secret-123",
	})
	if save.Code != 200 {
		t.Fatalf("save secret: expected 200, got %d: %s", save.Code, save.Body.String())
	}
	var saveResp struct {
		Ref string `json:"ref"`
	}
	if err := json.Unmarshal(save.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("unmarshal save response: %v", err)
	}
	if saveResp.Ref == "" {
		t.Fatal("expected ref in save response")
	}

	w := getJSON(handler, "/api/resolve/VK:TEMP:"+saveResp.Ref)
	if w.Code != 200 {
		t.Fatalf("resolve tracked ref: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	catalog, err := srv.db.GetSecretCatalogByRef("VK:TEMP:" + saveResp.Ref)
	if err != nil {
		t.Fatalf("GetSecretCatalogByRef: %v", err)
	}
	if catalog.LastRevealedAt == nil {
		t.Fatal("expected last_revealed_at to be populated after resolve")
	}

	rows, err := srv.db.ListAuditEvents("secret", "VK:TEMP:"+saveResp.Ref)
	if err != nil {
		t.Fatalf("ListAuditEvents: %v", err)
	}
	if len(rows) == 0 || rows[0].Action != "resolve" {
		t.Fatalf("expected resolve audit row, got %+v", rows)
	}
}
