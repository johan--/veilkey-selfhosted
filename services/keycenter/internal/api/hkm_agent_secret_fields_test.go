package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHKM_AgentSecretFieldsRequireActiveLocalOrExternalLifecycle(t *testing.T) {
	srv, handler := setupHKMServer(t)
	_, agentHash := registerMockAgent(t, srv, "test-service", nil, map[string]string{"API_KEY": "TEMP:deadbeef"})

	w := postJSON(handler, "/api/agents/"+agentHash+"/secrets/API_KEY/fields", map[string]interface{}{
		"fields": []map[string]string{
			{"key": "OTP", "type": "otp", "value": "123456"},
		},
	})
	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHKM_AgentSecretFieldsCRUDForExternal(t *testing.T) {
	srv, handler := setupHKMServer(t)
	_, agentHash := registerMockAgent(t, srv, "github-vault", nil, map[string]string{"GITHUB_KEY": "EXTERNAL:deadbeef"})

	saveSecret := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "GITHUB_KEY",
		"value": "ghp_example_token",
	})
	if saveSecret.Code != 200 {
		t.Fatalf("save secret: expected 200, got %d: %s", saveSecret.Code, saveSecret.Body.String())
	}

	save := postJSON(handler, "/api/agents/"+agentHash+"/secrets/GITHUB_KEY/fields", map[string]interface{}{
		"fields": []map[string]string{
			{"key": "LOGIN_ID", "type": "login", "value": "team@dalsoop.com"},
			{"key": "OTP", "type": "otp", "value": "123456"},
			{"key": "KEY_PASSWORD", "type": "password", "value": "hunter2"},
		},
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save fields: expected 200, got %d: %s", save.Code, save.Body.String())
	}

	getSecret := getJSON(handler, "/api/agents/"+agentHash+"/secrets/GITHUB_KEY")
	if getSecret.Code != http.StatusOK {
		t.Fatalf("get secret: expected 200, got %d: %s", getSecret.Code, getSecret.Body.String())
	}
	var secretResp struct {
		FieldsCount int `json:"fields_count"`
	}
	if err := json.Unmarshal(getSecret.Body.Bytes(), &secretResp); err != nil {
		t.Fatalf("decode secret response: %v", err)
	}
	if secretResp.FieldsCount != 3 {
		t.Fatalf("fields_count = %d, want 3", secretResp.FieldsCount)
	}

	getField := getJSON(handler, "/api/agents/"+agentHash+"/secrets/GITHUB_KEY/fields/OTP")
	if getField.Code != http.StatusOK {
		t.Fatalf("get field: expected 200, got %d: %s", getField.Code, getField.Body.String())
	}
	var fieldResp struct {
		Field string `json:"field"`
		Type  string `json:"type"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(getField.Body.Bytes(), &fieldResp); err != nil {
		t.Fatalf("decode field response: %v", err)
	}
	if fieldResp.Field != "OTP" || fieldResp.Type != "otp" || fieldResp.Value != "123456" {
		t.Fatalf("unexpected field response: %+v", fieldResp)
	}

	list := getJSON(handler, "/api/agents/"+agentHash+"/secrets")
	if list.Code != http.StatusOK {
		t.Fatalf("list secrets: expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		Secrets []struct {
			Name        string `json:"name"`
			FieldsCount int    `json:"fields_count"`
		} `json:"secrets"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list response: %v", err)
	}
	if len(listResp.Secrets) != 1 || listResp.Secrets[0].FieldsCount != 3 {
		t.Fatalf("unexpected list response: %+v", listResp.Secrets)
	}
}

func TestHKM_AgentSecretFieldsCRUDForLocal(t *testing.T) {
	srv, handler := setupHKMServer(t)
	_, agentHash := registerMockAgent(t, srv, "local-vault", nil, map[string]string{"INTERNAL_API_KEY": "deadbeef"})

	saveSecret := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "INTERNAL_API_KEY",
		"value": "lv_example_token",
	})
	if saveSecret.Code != http.StatusOK {
		t.Fatalf("save secret: expected 200, got %d: %s", saveSecret.Code, saveSecret.Body.String())
	}

	save := postJSON(handler, "/api/agents/"+agentHash+"/secrets/INTERNAL_API_KEY/fields", map[string]interface{}{
		"fields": []map[string]string{
			{"key": "LOGIN_ID", "type": "login", "value": "svc-internal"},
			{"key": "KEY_PASSWORD", "type": "password", "value": "local-pass"},
		},
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save fields: expected 200, got %d: %s", save.Code, save.Body.String())
	}

	var saveResp struct {
		Token string `json:"token"`
		Scope string `json:"scope"`
	}
	if err := json.Unmarshal(save.Body.Bytes(), &saveResp); err != nil {
		t.Fatalf("decode save response: %v", err)
	}
	if saveResp.Token != "VK:LOCAL:deadbeef" || saveResp.Scope != "LOCAL" {
		t.Fatalf("unexpected save response: %+v", saveResp)
	}

	getField := getJSON(handler, "/api/agents/"+agentHash+"/secrets/INTERNAL_API_KEY/fields/KEY_PASSWORD")
	if getField.Code != http.StatusOK {
		t.Fatalf("get field: expected 200, got %d: %s", getField.Code, getField.Body.String())
	}
}

func TestHKM_AgentSecretFieldsDelete(t *testing.T) {
	srv, handler := setupHKMServer(t)
	_, agentHash := registerMockAgent(t, srv, "delete-vault", nil, map[string]string{"GITHUB_KEY": "EXTERNAL:deadbeef"})

	saveSecret := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "GITHUB_KEY",
		"value": "ghp_example_token",
	})
	if saveSecret.Code != http.StatusOK {
		t.Fatalf("save secret: expected 200, got %d: %s", saveSecret.Code, saveSecret.Body.String())
	}
	save := postJSON(handler, "/api/agents/"+agentHash+"/secrets/GITHUB_KEY/fields", map[string]interface{}{
		"fields": []map[string]string{
			{"key": "OTP", "type": "otp", "value": "123456"},
		},
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save fields: expected 200, got %d: %s", save.Code, save.Body.String())
	}

	del := deleteJSON(handler, "/api/agents/"+agentHash+"/secrets/GITHUB_KEY/fields/OTP")
	if del.Code != http.StatusOK {
		t.Fatalf("delete field: expected 200, got %d: %s", del.Code, del.Body.String())
	}

	getField := getJSON(handler, "/api/agents/"+agentHash+"/secrets/GITHUB_KEY/fields/OTP")
	if getField.Code != http.StatusNotFound {
		t.Fatalf("expected 404 after delete, got %d: %s", getField.Code, getField.Body.String())
	}
}
