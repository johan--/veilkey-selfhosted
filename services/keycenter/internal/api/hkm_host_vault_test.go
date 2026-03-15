package api

import (
	"encoding/json"
	"net/http"
	"testing"

	"veilkey-keycenter/internal/crypto"
	"veilkey-keycenter/internal/db"
)

func TestHostVaultRoutesExposeLocalKeysAndConfigs(t *testing.T) {
	srv, handler := setupHKMServer(t)

	info, err := srv.db.GetNodeInfo()
	if err != nil {
		t.Fatalf("GetNodeInfo: %v", err)
	}
	localDEK, err := crypto.DecryptDEK(srv.kek, info.DEK, info.DEKNonce)
	if err != nil {
		t.Fatalf("DecryptDEK: %v", err)
	}
	ciphertext, nonce, err := crypto.Encrypt(localDEK, []byte("host-secret-value"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if err := srv.db.SaveSecret(&db.Secret{
		ID:         crypto.GenerateUUID(),
		Name:       "HOST_TEMP_KEY",
		Ref:        "deadbeef",
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Version:    info.Version,
	}); err != nil {
		t.Fatalf("SaveSecret: %v", err)
	}
	if err := srv.db.SaveRefWithName(db.RefParts{Family: "VK", Scope: "TEMP", ID: "deadbeef"}, "local-cipher", info.Version, "temp", "", "HOST_TEMP_KEY"); err != nil {
		t.Fatalf("SaveRefWithName(VK): %v", err)
	}
	if err := srv.db.SaveRefWithName(db.RefParts{Family: "VE", Scope: "LOCAL", ID: "HOST_URL"}, "local-config", info.Version, "active", "", "HOST_URL"); err != nil {
		t.Fatalf("SaveRefWithName(VE): %v", err)
	}

	keys := getJSON(handler, "/api/host-vault/keys")
	if keys.Code != http.StatusOK {
		t.Fatalf("keys code = %d body=%s", keys.Code, keys.Body.String())
	}
	var keysResp struct {
		Secrets []struct {
			Name  string `json:"name"`
			Scope string `json:"scope"`
			Token string `json:"token"`
		} `json:"secrets"`
	}
	if err := json.Unmarshal(keys.Body.Bytes(), &keysResp); err != nil {
		t.Fatalf("unmarshal keys: %v", err)
	}
	if len(keysResp.Secrets) != 1 || keysResp.Secrets[0].Name != "HOST_TEMP_KEY" {
		t.Fatalf("unexpected keys payload: %s", keys.Body.String())
	}
	if keysResp.Secrets[0].Scope != "TEMP" {
		t.Fatalf("key scope = %q", keysResp.Secrets[0].Scope)
	}

	keyDetail := getJSON(handler, "/api/host-vault/keys/HOST_TEMP_KEY")
	if keyDetail.Code != http.StatusOK {
		t.Fatalf("key detail code = %d body=%s", keyDetail.Code, keyDetail.Body.String())
	}
	var keyResp struct {
		Name  string `json:"name"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(keyDetail.Body.Bytes(), &keyResp); err != nil {
		t.Fatalf("unmarshal key detail: %v", err)
	}
	if keyResp.Name != "HOST_TEMP_KEY" || keyResp.Value != "host-secret-value" {
		t.Fatalf("unexpected key detail: %s", keyDetail.Body.String())
	}

	configs := getJSON(handler, "/api/host-vault/configs")
	if configs.Code != http.StatusOK {
		t.Fatalf("configs code = %d body=%s", configs.Code, configs.Body.String())
	}
	var cfgResp struct {
		Configs []struct {
			Key   string `json:"key"`
			Scope string `json:"scope"`
		} `json:"configs"`
	}
	if err := json.Unmarshal(configs.Body.Bytes(), &cfgResp); err != nil {
		t.Fatalf("unmarshal configs: %v", err)
	}
	if len(cfgResp.Configs) != 1 || cfgResp.Configs[0].Key != "HOST_URL" || cfgResp.Configs[0].Scope != "LOCAL" {
		t.Fatalf("unexpected configs payload: %s", configs.Body.String())
	}
}

func TestHostVaultSaveEndpointsPersistScopes(t *testing.T) {
	srv, handler := setupHKMServer(t)

	saveKey := postJSON(handler, "/api/host-vault/keys", map[string]any{
		"name":  "HOST_PROMOTE_KEY",
		"value": "copied-secret",
		"scope": "TEMP",
	})
	if saveKey.Code != http.StatusOK {
		t.Fatalf("save host key expected 200, got %d: %s", saveKey.Code, saveKey.Body.String())
	}
	keyDetail := getJSON(handler, "/api/host-vault/keys/HOST_PROMOTE_KEY")
	if keyDetail.Code != http.StatusOK {
		t.Fatalf("host key detail expected 200, got %d: %s", keyDetail.Code, keyDetail.Body.String())
	}
	var keyResp struct {
		Scope string `json:"scope"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(keyDetail.Body.Bytes(), &keyResp); err != nil {
		t.Fatalf("decode host key detail: %v", err)
	}
	if keyResp.Scope != "TEMP" || keyResp.Value != "copied-secret" {
		t.Fatalf("unexpected host key detail: %s", keyDetail.Body.String())
	}

	saveCfg := postJSON(handler, "/api/host-vault/configs", map[string]any{
		"key":   "HOST_ENDPOINT",
		"value": "https://host.example",
		"scope": "EXTERNAL",
	})
	if saveCfg.Code != http.StatusOK {
		t.Fatalf("save host config expected 200, got %d: %s", saveCfg.Code, saveCfg.Body.String())
	}
	cfgDetail := getJSON(handler, "/api/host-vault/configs/HOST_ENDPOINT")
	if cfgDetail.Code != http.StatusOK {
		t.Fatalf("host config detail expected 200, got %d: %s", cfgDetail.Code, cfgDetail.Body.String())
	}
	var cfgResp struct {
		Scope string `json:"scope"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(cfgDetail.Body.Bytes(), &cfgResp); err != nil {
		t.Fatalf("decode host config detail: %v", err)
	}
	if cfgResp.Scope != "EXTERNAL" || cfgResp.Value != "https://host.example" {
		t.Fatalf("unexpected host config detail: %s", cfgDetail.Body.String())
	}
	_ = srv
}

func TestHostVaultKeysCollapseDuplicateRefsBySecretName(t *testing.T) {
	srv, handler := setupHKMServer(t)

	info, err := srv.db.GetNodeInfo()
	if err != nil {
		t.Fatalf("GetNodeInfo: %v", err)
	}
	localDEK, err := crypto.DecryptDEK(srv.kek, info.DEK, info.DEKNonce)
	if err != nil {
		t.Fatalf("DecryptDEK: %v", err)
	}
	ciphertext, nonce, err := crypto.Encrypt(localDEK, []byte("deduped-secret"))
	if err != nil {
		t.Fatalf("Encrypt: %v", err)
	}
	if err := srv.db.SaveSecret(&db.Secret{
		ID:         crypto.GenerateUUID(),
		Name:       "DUPLICATE_KEY",
		Ref:        "hostdup01",
		Ciphertext: ciphertext,
		Nonce:      nonce,
		Version:    info.Version,
	}); err != nil {
		t.Fatalf("SaveSecret: %v", err)
	}
	if err := srv.db.SaveRefWithName(db.RefParts{Family: "VK", Scope: "TEMP", ID: "hostdup01"}, "local-cipher", info.Version, "temp", "", "DUPLICATE_KEY"); err != nil {
		t.Fatalf("SaveRefWithName(temp): %v", err)
	}
	if err := srv.db.SaveRefWithName(db.RefParts{Family: "VK", Scope: "LOCAL", ID: "hostdup01"}, "local-cipher", info.Version, "active", "", "DUPLICATE_KEY"); err != nil {
		t.Fatalf("SaveRefWithName(local): %v", err)
	}

	keys := getJSON(handler, "/api/host-vault/keys")
	if keys.Code != http.StatusOK {
		t.Fatalf("keys code = %d body=%s", keys.Code, keys.Body.String())
	}
	var keysResp struct {
		Secrets []struct {
			Name  string `json:"name"`
			Scope string `json:"scope"`
			Token string `json:"token"`
		} `json:"secrets"`
	}
	if err := json.Unmarshal(keys.Body.Bytes(), &keysResp); err != nil {
		t.Fatalf("unmarshal keys: %v", err)
	}
	if len(keysResp.Secrets) != 1 {
		t.Fatalf("expected collapsed host key list, got %s", keys.Body.String())
	}
	if keysResp.Secrets[0].Name != "DUPLICATE_KEY" || keysResp.Secrets[0].Scope != "LOCAL" || keysResp.Secrets[0].Token != "VK:LOCAL:hostdup01" {
		t.Fatalf("unexpected collapsed host key payload: %s", keys.Body.String())
	}

	keyDetail := getJSON(handler, "/api/host-vault/keys/DUPLICATE_KEY")
	if keyDetail.Code != http.StatusOK {
		t.Fatalf("key detail code = %d body=%s", keyDetail.Code, keyDetail.Body.String())
	}
	var keyResp struct {
		Scope string `json:"scope"`
		Token string `json:"token"`
		Value string `json:"value"`
	}
	if err := json.Unmarshal(keyDetail.Body.Bytes(), &keyResp); err != nil {
		t.Fatalf("unmarshal key detail: %v", err)
	}
	if keyResp.Scope != "LOCAL" || keyResp.Token != "VK:LOCAL:hostdup01" || keyResp.Value != "deduped-secret" {
		t.Fatalf("unexpected key detail payload: %s", keyDetail.Body.String())
	}
}
