package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestRefPolicyEndpoint(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/ref-policy")
	if w.Code != http.StatusOK {
		t.Fatalf("ref policy: expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Count    int `json:"count"`
		Policies []struct {
			Family        string            `json:"family"`
			DefaultScope  string            `json:"default_scope"`
			AllowedScopes []string          `json:"allowed_scopes"`
			DefaultStatus map[string]string `json:"default_statuses"`
		} `json:"policies"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("unmarshal ref policy: %v", err)
	}
	if resp.Count != 2 {
		t.Fatalf("count = %d, want 2", resp.Count)
	}
	if len(resp.Policies) != 2 {
		t.Fatalf("policies len = %d, want 2", len(resp.Policies))
	}
	for _, policy := range resp.Policies {
		if policy.DefaultScope != "TEMP" {
			t.Fatalf("%s default_scope = %q, want TEMP", policy.Family, policy.DefaultScope)
		}
		if len(policy.AllowedScopes) != 3 {
			t.Fatalf("%s allowed_scopes len = %d, want 3", policy.Family, len(policy.AllowedScopes))
		}
		if policy.DefaultStatus["TEMP"] != "temp" {
			t.Fatalf("%s TEMP status = %q, want temp", policy.Family, policy.DefaultStatus["TEMP"])
		}
		if policy.DefaultStatus["LOCAL"] != "active" || policy.DefaultStatus["EXTERNAL"] != "active" {
			t.Fatalf("%s active defaults missing: %+v", policy.Family, policy.DefaultStatus)
		}
	}
}
