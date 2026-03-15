package api

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"veilkey-localvault/internal/db"
)

func TestHandleSaveSecretFieldsRequiresActiveLocalOrExternalSecret(t *testing.T) {
	server := setupReencryptTestServer(t)
	if err := server.db.UpdateSecretLifecycle("deadbeef", "TEMP", "temp"); err != nil {
		t.Fatalf("UpdateSecretLifecycle: %v", err)
	}
	handler := server.SetupRoutes()

	req := httptest.NewRequest(http.MethodPost, "/api/secrets/fields", bytes.NewBufferString(`{"name":"TEST_SECRET","fields":[{"key":"OTP","type":"otp","ciphertext":"AQID","nonce":"AQIDBAUGBwgJCgsM"}]}`))
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusConflict {
		t.Fatalf("expected 409, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleSaveSecretFieldsStoresMetadataAndCipherForExternal(t *testing.T) {
	server := setupReencryptTestServer(t)
	if err := server.db.UpdateSecretLifecycle("deadbeef", "EXTERNAL", "active"); err != nil {
		t.Fatalf("UpdateSecretLifecycle: %v", err)
	}
	handler := server.SetupRoutes()

	req := httptest.NewRequest(http.MethodPost, "/api/secrets/fields", bytes.NewBufferString(`{"name":"TEST_SECRET","fields":[{"key":"LOGIN_ID","type":"login","ciphertext":"AQID","nonce":"AQIDBAUGBwgJCgsM"},{"key":"OTP","type":"otp","ciphertext":"BAUG","nonce":"AQIDBAUGBwgJCgsM"}]}`))
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	metaReq := httptest.NewRequest(http.MethodGet, "/api/secrets/meta/TEST_SECRET", nil)
	metaW := httptest.NewRecorder()
	handler.ServeHTTP(metaW, metaReq)
	if metaW.Code != http.StatusOK {
		t.Fatalf("meta expected 200, got %d: %s", metaW.Code, metaW.Body.String())
	}

	var metaResp struct {
		DisplayName    string  `json:"display_name"`
		Class          string  `json:"class"`
		LastRotatedAt  *string `json:"last_rotated_at"`
		LastRevealedAt *string `json:"last_revealed_at"`
		Fields         []struct {
			Key             string `json:"key"`
			Type            string `json:"type"`
			FieldRole       string `json:"field_role"`
			DisplayName     string `json:"display_name"`
			MaskedByDefault bool   `json:"masked_by_default"`
			Required        bool   `json:"required"`
			SortOrder       int    `json:"sort_order"`
		} `json:"fields"`
		FieldsCount int `json:"fields_count"`
	}
	if err := json.Unmarshal(metaW.Body.Bytes(), &metaResp); err != nil {
		t.Fatalf("decode meta response: %v", err)
	}
	if metaResp.FieldsCount != 2 {
		t.Fatalf("fields_count = %d, want 2", metaResp.FieldsCount)
	}
	if metaResp.DisplayName != "TEST_SECRET" {
		t.Fatalf("display_name = %q, want TEST_SECRET", metaResp.DisplayName)
	}
	if metaResp.Class != "key" {
		t.Fatalf("class = %q, want key", metaResp.Class)
	}
	if metaResp.LastRotatedAt != nil || metaResp.LastRevealedAt != nil {
		t.Fatalf("expected null timestamps by default, got rotated=%v revealed=%v", metaResp.LastRotatedAt, metaResp.LastRevealedAt)
	}
	if metaResp.Fields[1].FieldRole != "otp" {
		t.Fatalf("field_role = %q, want otp", metaResp.Fields[1].FieldRole)
	}
	if metaResp.Fields[1].DisplayName != "OTP" {
		t.Fatalf("field display_name = %q, want OTP", metaResp.Fields[1].DisplayName)
	}
	if metaResp.Fields[1].MaskedByDefault {
		t.Fatal("masked_by_default should default to false in current save path")
	}

	cipherReq := httptest.NewRequest(http.MethodGet, "/api/cipher/deadbeef/fields/OTP", nil)
	cipherReq.RemoteAddr = "127.0.0.1:12345"
	cipherW := httptest.NewRecorder()
	handler.ServeHTTP(cipherW, cipherReq)
	if cipherW.Code != http.StatusOK {
		t.Fatalf("cipher field expected 200, got %d: %s", cipherW.Code, cipherW.Body.String())
	}

	var cipherResp struct {
		Field string `json:"field"`
		Type  string `json:"type"`
	}
	if err := json.Unmarshal(cipherW.Body.Bytes(), &cipherResp); err != nil {
		t.Fatalf("decode cipher field response: %v", err)
	}
	if cipherResp.Field != "OTP" || cipherResp.Type != "otp" {
		t.Fatalf("unexpected cipher field response: %+v", cipherResp)
	}
}

func TestHandleSaveSecretFieldsStoresMetadataAndCipherForLocal(t *testing.T) {
	server := setupReencryptTestServer(t)
	if err := server.db.UpdateSecretLifecycle("deadbeef", "LOCAL", "active"); err != nil {
		t.Fatalf("UpdateSecretLifecycle: %v", err)
	}
	handler := server.SetupRoutes()

	req := httptest.NewRequest(http.MethodPost, "/api/secrets/fields", bytes.NewBufferString(`{"name":"TEST_SECRET","fields":[{"key":"KEY_PASSWORD","type":"password","ciphertext":"AQID","nonce":"AQIDBAUGBwgJCgsM"}]}`))
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestHandleDeleteSecretFieldRemovesStoredField(t *testing.T) {
	server := setupReencryptTestServer(t)
	if err := server.db.UpdateSecretLifecycle("deadbeef", "LOCAL", "active"); err != nil {
		t.Fatalf("UpdateSecretLifecycle: %v", err)
	}
	if err := server.db.SaveSecretFields("TEST_SECRET", []db.SecretField{{
		SecretName: "TEST_SECRET",
		FieldKey:   "OTP",
		FieldType:  "otp",
		Ciphertext: []byte{1, 2, 3},
		Nonce:      []byte("123456789012"),
	}}); err != nil {
		t.Fatalf("SaveSecretFields: %v", err)
	}
	handler := server.SetupRoutes()

	req := httptest.NewRequest(http.MethodDelete, "/api/secrets/TEST_SECRET/fields/OTP", nil)
	req.RemoteAddr = "127.0.0.1:12345"
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
	if _, err := server.db.GetSecretField("TEST_SECRET", "OTP"); err == nil {
		t.Fatal("expected field to be deleted")
	}
}
