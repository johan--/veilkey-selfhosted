package api

import (
	"strings"
	"testing"
)

func TestHKM_AgentSecretsRejectSpecialCharacterName(t *testing.T) {
	srv, handler := setupHKMServer(t)
	_, agentHash := registerMockAgent(t, srv, "reject-secret-agent", map[string]string{}, nil)

	w := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "BAD@KEY",
		"value": "super-secret-123",
	})
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "name must match") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHKM_AgentSecretsRejectLowercaseName(t *testing.T) {
	srv, handler := setupHKMServer(t)
	_, agentHash := registerMockAgent(t, srv, "reject-secret-agent-lower", map[string]string{}, nil)

	w := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "Bad_Key",
		"value": "super-secret-123",
	})
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "name must match") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHKM_AgentConfigsRejectSpecialCharacterKey(t *testing.T) {
	srv, handler := setupHKMServer(t)
	_, agentHash := registerMockAgent(t, srv, "reject-config-agent", map[string]string{}, nil)

	w := postJSON(handler, "/api/agents/"+agentHash+"/configs", map[string]string{
		"key":   "BAD@KEY",
		"value": "x",
	})
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "key must match") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHKM_AgentConfigsRejectLowercaseKey(t *testing.T) {
	srv, handler := setupHKMServer(t)
	_, agentHash := registerMockAgent(t, srv, "reject-config-agent-lower", map[string]string{}, nil)

	w := postJSON(handler, "/api/agents/"+agentHash+"/configs", map[string]string{
		"key":   "Bad_Key",
		"value": "x",
	})
	if w.Code != 400 {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "key must match") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
