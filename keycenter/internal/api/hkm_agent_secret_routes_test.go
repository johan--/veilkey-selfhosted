package api

import (
	"net/http"
	"testing"
)

func TestHKM_AgentSecretRoutesRemainSupportedWhileDirectSecretsStayRemoved(t *testing.T) {
	srv, handler := setupHKMServer(t)

	_, agentHash := registerMockAgent(t, srv, "route-contract", map[string]string{}, nil)

	directList := getJSON(handler, "/api/secrets")
	if directList.Code != http.StatusNotFound {
		t.Fatalf("direct secret list expected 404, got %d", directList.Code)
	}

	directSave := postJSON(handler, "/api/secrets", map[string]string{
		"name":  "API_KEY",
		"value": "should-not-work",
	})
	if directSave.Code != http.StatusNotFound {
		t.Fatalf("direct secret save expected 404, got %d", directSave.Code)
	}

	agentSave := postJSON(handler, "/api/agents/"+agentHash+"/secrets", map[string]string{
		"name":  "API_KEY",
		"value": "supported-path",
	})
	if agentSave.Code != http.StatusOK {
		t.Fatalf("agent secret save expected 200, got %d: %s", agentSave.Code, agentSave.Body.String())
	}

	agentList := getJSON(handler, "/api/agents/"+agentHash+"/secrets")
	if agentList.Code != http.StatusOK {
		t.Fatalf("agent secret list expected 200, got %d: %s", agentList.Code, agentList.Body.String())
	}

	agentGet := getJSON(handler, "/api/agents/"+agentHash+"/secrets/API_KEY")
	if agentGet.Code != http.StatusOK {
		t.Fatalf("agent secret get expected 200, got %d: %s", agentGet.Code, agentGet.Body.String())
	}
}
