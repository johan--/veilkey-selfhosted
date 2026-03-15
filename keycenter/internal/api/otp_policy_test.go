package api

import (
	"encoding/json"
	"testing"
)

func TestOTPPolicyReturnsExemptList(t *testing.T) {
	_, handler, _ := setupServerWithPassword(t, "install-pass")

	resp := getJSON(handler, "/api/otp-policy")
	if resp.Code != 200 {
		t.Fatalf("expected 200, got %d: %s", resp.Code, resp.Body.String())
	}

	var body struct {
		Exempt []string `json:"exempt"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	expected := map[string]bool{
		"secret reveal":       false,
		"secret field reveal": false,
	}
	for _, op := range body.Exempt {
		if _, ok := expected[op]; ok {
			expected[op] = true
		}
	}
	for op, found := range expected {
		if !found {
			t.Errorf("expected %q in exempt list", op)
		}
	}
}
