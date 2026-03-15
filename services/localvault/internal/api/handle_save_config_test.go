package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleSaveConfigRestoresOperationalLifecycleOnOverwrite(t *testing.T) {
	server := setupReencryptTestServer(t)

	if err := server.db.SaveConfig("APP_URL", "https://old.example.test"); err != nil {
		t.Fatalf("SaveConfig seed: %v", err)
	}
	if err := server.db.UpdateConfigLifecycle("APP_URL", "EXTERNAL", "block"); err != nil {
		t.Fatalf("UpdateConfigLifecycle seed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/configs", bytes.NewBufferString(`{"key":"APP_URL","value":"https://new.example.test"}`))
	w := httptest.NewRecorder()
	server.handleSaveConfig(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Key    string `json:"key"`
		Value  string `json:"value"`
		Ref    string `json:"ref"`
		Scope  string `json:"scope"`
		Status string `json:"status"`
		Action string `json:"action"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Key != "APP_URL" || resp.Value != "https://new.example.test" {
		t.Fatalf("response = %+v", resp)
	}
	if resp.Ref != "VE:LOCAL:APP_URL" || resp.Scope != "LOCAL" || resp.Status != "active" || resp.Action != "saved" {
		t.Fatalf("response lifecycle = %+v", resp)
	}

	config, err := server.db.GetConfig("APP_URL")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if config.Scope != "LOCAL" || config.Status != "active" {
		t.Fatalf("stored lifecycle = %s/%s", config.Scope, config.Status)
	}
}

func TestHandleSaveConfigsBulkRestoresOperationalLifecycleOnOverwrite(t *testing.T) {
	server := setupReencryptTestServer(t)

	if err := server.db.SaveConfig("APP_ENV", "staging"); err != nil {
		t.Fatalf("SaveConfig seed: %v", err)
	}
	if err := server.db.UpdateConfigLifecycle("APP_ENV", "EXTERNAL", "archive"); err != nil {
		t.Fatalf("UpdateConfigLifecycle seed: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/api/configs/bulk", bytes.NewBufferString(`{"configs":{"APP_ENV":"production"}}`))
	w := httptest.NewRecorder()
	server.handleSaveConfigsBulk(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Saved int `json:"saved"`
	}
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.Saved != 1 {
		t.Fatalf("saved = %d", resp.Saved)
	}

	config, err := server.db.GetConfig("APP_ENV")
	if err != nil {
		t.Fatalf("GetConfig: %v", err)
	}
	if config.Value != "production" {
		t.Fatalf("value = %q", config.Value)
	}
	if config.Scope != "LOCAL" || config.Status != "active" {
		t.Fatalf("stored lifecycle = %s/%s", config.Scope, config.Status)
	}
}
