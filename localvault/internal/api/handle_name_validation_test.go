package api

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandleSaveConfigRejectsSpecialCharacterKey(t *testing.T) {
	server := setupReencryptTestServer(t)

	req := httptest.NewRequest(http.MethodPut, "/api/configs", bytes.NewBufferString(`{"key":"BAD@KEY","value":"x"}`))
	w := httptest.NewRecorder()
	server.handleSaveConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "key must match") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandleSaveConfigRejectsLowercaseKey(t *testing.T) {
	server := setupReencryptTestServer(t)

	req := httptest.NewRequest(http.MethodPut, "/api/configs", bytes.NewBufferString(`{"key":"Bad_Key","value":"x"}`))
	w := httptest.NewRecorder()
	server.handleSaveConfig(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "key must match") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandleSaveCipherRejectsSpecialCharacterName(t *testing.T) {
	server := setupReencryptTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/cipher", bytes.NewBufferString(`{"name":"BAD@KEY","ciphertext":"AQID","nonce":"AQIDBAUGBwgJCgsM","version":1}`))
	w := httptest.NewRecorder()
	server.handleSaveCipher(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "name must match") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}

func TestHandleSaveCipherRejectsLowercaseName(t *testing.T) {
	server := setupReencryptTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/cipher", bytes.NewBufferString(`{"name":"Bad_Key","ciphertext":"AQID","nonce":"AQIDBAUGBwgJCgsM","version":1}`))
	w := httptest.NewRecorder()
	server.handleSaveCipher(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", w.Code, w.Body.String())
	}
	if !strings.Contains(w.Body.String(), "name must match") {
		t.Fatalf("unexpected body: %s", w.Body.String())
	}
}
