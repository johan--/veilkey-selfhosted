package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleSaveCipherStoresTempLifecycle(t *testing.T) {
	server := setupReencryptTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/cipher", bytes.NewBufferString(`{"name":"API_KEY","ciphertext":"AQID","nonce":"AQIDBAUGBwgJCgsM","version":1}`))
	w := httptest.NewRecorder()
	server.handleSaveCipher(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Ref    string `json:"ref"`
		Token  string `json:"token"`
		Scope  string `json:"scope"`
		Status string `json:"status"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Scope != "TEMP" || resp.Status != "temp" {
		t.Fatalf("unexpected lifecycle: scope=%q status=%q", resp.Scope, resp.Status)
	}
	if resp.Token != "VK:TEMP:"+resp.Ref {
		t.Fatalf("token = %q, want VK:TEMP:%s", resp.Token, resp.Ref)
	}
	if resp.Action != "created" {
		t.Fatalf("action = %q, want created", resp.Action)
	}

	secret, err := server.db.GetSecretByName("API_KEY")
	if err != nil {
		t.Fatalf("GetSecretByName: %v", err)
	}
	if secret.Scope != "TEMP" || secret.Status != "temp" {
		t.Fatalf("stored lifecycle: scope=%q status=%q", secret.Scope, secret.Status)
	}
}

func TestHandleSaveCipherPreservesExistingLifecycleOnUpdate(t *testing.T) {
	server := setupReencryptTestServer(t)

	if err := server.db.UpdateSecretLifecycle("deadbeef", "LOCAL", "active"); err != nil {
		t.Fatalf("UpdateSecretLifecycle: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/api/cipher", bytes.NewBufferString(`{"name":"TEST_SECRET","ciphertext":"AQID","nonce":"AQIDBAUGBwgJCgsM","version":1}`))
	w := httptest.NewRecorder()
	server.handleSaveCipher(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Ref    string `json:"ref"`
		Token  string `json:"token"`
		Scope  string `json:"scope"`
		Status string `json:"status"`
		Action string `json:"action"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Scope != "LOCAL" || resp.Status != "active" {
		t.Fatalf("unexpected lifecycle: scope=%q status=%q", resp.Scope, resp.Status)
	}
	if resp.Token != "VK:LOCAL:"+resp.Ref {
		t.Fatalf("token = %q, want VK:LOCAL:%s", resp.Token, resp.Ref)
	}
	if resp.Action != "updated" {
		t.Fatalf("action = %q, want updated", resp.Action)
	}
}

func TestHandleCipherMarksSecretRevealed(t *testing.T) {
	server := setupReencryptTestServer(t)
	handler := server.SetupRoutes()

	req := httptest.NewRequest(http.MethodGet, "/api/cipher/deadbeef", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	secret, err := server.db.GetSecretByRef("deadbeef")
	if err != nil {
		t.Fatalf("GetSecretByRef: %v", err)
	}
	if !secret.LastRevealedAt.Valid {
		t.Fatal("expected last_revealed_at to be set after cipher read")
	}

	metaReq := httptest.NewRequest(http.MethodGet, "/api/secrets/meta/TEST_SECRET", nil)
	metaW := httptest.NewRecorder()
	handler.ServeHTTP(metaW, metaReq)

	if metaW.Code != http.StatusOK {
		t.Fatalf("meta expected 200, got %d: %s", metaW.Code, metaW.Body.String())
	}

	var metaResp struct {
		LastRevealedAt *string `json:"last_revealed_at"`
	}
	if err := json.Unmarshal(metaW.Body.Bytes(), &metaResp); err != nil {
		t.Fatalf("decode meta response: %v", err)
	}
	if metaResp.LastRevealedAt == nil {
		t.Fatal("expected serialized last_revealed_at to be present")
	}
	if _, err := time.Parse(time.RFC3339, *metaResp.LastRevealedAt); err != nil {
		t.Fatalf("last_revealed_at should be RFC3339, got %q", *metaResp.LastRevealedAt)
	}
}
