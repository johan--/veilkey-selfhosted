package api

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestEmailOTPChallengeRoundTrip(t *testing.T) {
	inbox := filepath.Join(t.TempDir(), "mail.txt")
	sendmail := filepath.Join(t.TempDir(), "sendmail")
	if err := os.WriteFile(sendmail, []byte("#!/bin/sh\ncat >"+inbox+"\n"), 0o755); err != nil {
		t.Fatalf("write fake sendmail: %v", err)
	}
	t.Setenv("VEILKEY_OTP_SENDMAIL", sendmail)
	t.Setenv("VEILKEY_OTP_SMTP_HOST", "")

	_, handler, _ := setupServerWithPassword(t, "install-pass")

	request := postJSON(handler, "/api/approvals/email-otp/request", map[string]any{
		"email":    "tex02@naver.com",
		"reason":   "manual send",
		"base_url": "https://keycenter.test",
	})
	if request.Code != 201 {
		t.Fatalf("create email otp challenge: expected 201, got %d: %s", request.Code, request.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(request.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	token, _ := resp["token"].(string)
	link, _ := resp["link"].(string)
	if token == "" || !strings.Contains(link, token) {
		t.Fatalf("expected tokenized link, got token=%q link=%q", token, link)
	}

	mailBody, err := os.ReadFile(inbox)
	if err != nil {
		t.Fatalf("read inbox: %v", err)
	}
	if !strings.Contains(string(mailBody), link) {
		t.Fatalf("expected approval link in mail, got %q", string(mailBody))
	}

	page := getJSON(handler, "/ui/approvals/email-otp?token="+token)
	if page.Code != 200 {
		t.Fatalf("email otp page: expected 200, got %d: %s", page.Code, page.Body.String())
	}
	if !strings.Contains(page.Body.String(), "Send code by email") {
		t.Fatalf("expected email otp page, got %q", page.Body.String())
	}

	sendCode := postForm(handler, "/ui/approvals/email-otp", map[string]string{
		"token":  token,
		"action": "send-code",
	})
	if sendCode.Code != 200 {
		t.Fatalf("send-code: expected 200, got %d: %s", sendCode.Code, sendCode.Body.String())
	}

	state := getJSON(handler, "/api/approvals/email-otp/state?token="+token)
	if state.Code != 200 {
		t.Fatalf("state: expected 200, got %d: %s", state.Code, state.Body.String())
	}
	if !strings.Contains(state.Body.String(), `"status":"pending"`) {
		t.Fatalf("expected pending state, got %q", state.Body.String())
	}
}
