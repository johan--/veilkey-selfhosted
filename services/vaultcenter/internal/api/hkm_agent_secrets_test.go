package api

import (
	"encoding/json"
	"testing"
)

func TestHKM_AgentSecretsSaveAndGet(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "secret-agent", map[string]string{}, nil)

	w := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "API_KEY",
		"value": "super-secret-123",
	})
	if w.Code != 200 {
		t.Fatalf("save secret: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var saveResp struct {
		Name   string `json:"name"`
		Ref    string `json:"ref"`
		Token  string `json:"token"`
		Scope  string `json:"scope"`
		Status string `json:"status"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("unmarshal save response: %v", err)
	}
	if saveResp.Name != "API_KEY" {
		t.Fatalf("name = %q", saveResp.Name)
	}
	if saveResp.Ref == "" {
		t.Fatal("ref should not be empty")
	}
	if saveResp.Token != "VK:TEMP:"+saveResp.Ref {
		t.Fatalf("token = %q", saveResp.Token)
	}
	if saveResp.Scope != "TEMP" {
		t.Fatalf("scope = %q", saveResp.Scope)
	}
	if saveResp.Status != "temp" {
		t.Fatalf("status = %q", saveResp.Status)
	}
	if _, err := srv.db.GetRef("VK:TEMP:" + saveResp.Ref); err != nil {
		t.Fatalf("tracked ref missing after save: %v", err)
	}
	saveAudit, err := srv.db.ListAuditEvents("secret", "VK:TEMP:"+saveResp.Ref)
	if err != nil {
		t.Fatalf("ListAuditEvents(save): %v", err)
	}
	if len(saveAudit) == 0 || saveAudit[0].Action != "save" {
		t.Fatalf("expected secret save audit, got %+v", saveAudit)
	}

	w = getJSON(handler, "/api/agents/"+agentHash+"/secrets/API_KEY")
	if w.Code != 200 {
		t.Fatalf("get secret: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var getResp struct {
		Name  string `json:"name"`
		Value string `json:"value"`
		Ref   string `json:"ref"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("unmarshal get response: %v", err)
	}
	if getResp.Name != "API_KEY" {
		t.Fatalf("name = %q", getResp.Name)
	}
	if getResp.Value != "super-secret-123" {
		t.Fatalf("value = %q", getResp.Value)
	}
	if getResp.Ref != saveResp.Ref {
		t.Fatalf("ref = %q, want %q", getResp.Ref, saveResp.Ref)
	}

	w = deleteJSON(handler, "/api/agents/"+agentHash+"/secrets/API_KEY")
	if w.Code != 200 {
		t.Fatalf("delete secret: expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if _, err := srv.db.GetRef("VK:TEMP:" + saveResp.Ref); err == nil {
		t.Fatal("tracked ref should be removed after delete")
	}
	deleteAudit, err := srv.db.ListAuditEvents("secret", "VK:TEMP:"+saveResp.Ref)
	if err != nil {
		t.Fatalf("ListAuditEvents(delete): %v", err)
	}
	if len(deleteAudit) < 2 || deleteAudit[0].Action != "delete" {
		t.Fatalf("expected secret delete audit, got %+v", deleteAudit)
	}
}

func TestHKM_AgentSecretsPublicResponsesAreCanonical(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "canonical-secret-agent", map[string]string{}, nil)

	for _, tc := range []struct {
		name  string
		value string
	}{
		{name: "LOCAL_ONLY", value: "local-only-secret"},
		{name: "ANOTHER_KEY", value: "another-secret"},
	} {
		save := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
			"name":  tc.name,
			"value": tc.value,
		})
		if save.Code != 200 {
			t.Fatalf("save %s: expected 200, got %d: %s", tc.name, save.Code, save.Body.String())
		}
	}

	list := getJSON(handler, "/api/agents/"+agentHash+"/secrets")
	if list.Code != 200 {
		t.Fatalf("list secrets: expected 200, got %d: %s", list.Code, list.Body.String())
	}

	var listResp struct {
		Secrets []map[string]interface{} `json:"secrets"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("unmarshal list response: %v", err)
	}
	if len(listResp.Secrets) != 2 {
		t.Fatalf("secrets len = %d, want 2", len(listResp.Secrets))
	}
	for _, sec := range listResp.Secrets {
		ref, _ := sec["ref"].(string)
		token, _ := sec["token"].(string)
		scope, _ := sec["scope"].(string)
		status, _ := sec["status"].(string)
		if ref == "" {
			t.Fatalf("list secret missing ref: %#v", sec)
		}
		if token != "VK:"+scope+":"+ref {
			t.Fatalf("token = %q, want canonical VK:%s:%s", token, scope, ref)
		}
		if status == "" {
			t.Fatalf("status should not be empty: %#v", sec)
		}
		if _, ok := sec["vault_token"]; ok {
			t.Fatalf("legacy vault_token should not be exposed: %#v", sec)
		}
	}

	get := getJSON(handler, "/api/agents/"+agentHash+"/secrets/LOCAL_ONLY")
	if get.Code != 200 {
		t.Fatalf("get secret: expected 200, got %d: %s", get.Code, get.Body.String())
	}

	var getResp map[string]interface{}
	if err := json.Unmarshal(get.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("unmarshal get response: %v", err)
	}
	ref, _ := getResp["ref"].(string)
	token, _ := getResp["token"].(string)
	scope, _ := getResp["scope"].(string)
	status, _ := getResp["status"].(string)
	if ref == "" {
		t.Fatalf("get response missing ref: %#v", getResp)
	}
	if token != "VK:"+scope+":"+ref {
		t.Fatalf("token = %q, want canonical VK:%s:%s", token, scope, ref)
	}
	if scope != "TEMP" {
		t.Fatalf("scope = %q, want TEMP", scope)
	}
	if status != "temp" {
		t.Fatalf("status = %q, want temp", status)
	}
	if _, ok := getResp["vault_token"]; ok {
		t.Fatalf("legacy vault_token should not be exposed: %#v", getResp)
	}
}

func TestHKM_AgentSecretsSavePreservesReturnedLifecycle(t *testing.T) {
	srv, handler := setupHKMServer(t)

	configs := map[string]string{}
	secrets := map[string]string{"API_KEY": "deadbeef"}
	_, agentHash := registerMockAgent(t, srv, "lifecycle-agent", configs, secrets)

	w := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "API_KEY",
		"value": "super-secret-123",
	})
	if w.Code != 200 {
		t.Fatalf("save secret: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Ref    string `json:"ref"`
		Token  string `json:"token"`
		Scope  string `json:"scope"`
		Status string `json:"status"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal save response: %v", err)
	}
	if resp.Scope != "LOCAL" {
		t.Fatalf("scope = %q, want LOCAL", resp.Scope)
	}
	if resp.Status != "active" {
		t.Fatalf("status = %q, want active", resp.Status)
	}
	if resp.Token != "VK:LOCAL:"+resp.Ref {
		t.Fatalf("token = %q, want VK:LOCAL:%s", resp.Token, resp.Ref)
	}
	if resp.Action != "updated" {
		t.Fatalf("action = %q, want updated", resp.Action)
	}
	if _, err := srv.db.GetRef("VK:LOCAL:" + resp.Ref); err != nil {
		t.Fatalf("tracked ref missing after save: %v", err)
	}
}
