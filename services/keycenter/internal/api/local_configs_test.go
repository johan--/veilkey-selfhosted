package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLocalConfigsCRUD(t *testing.T) {
	_, handler := setupTestServer(t)

	save := postJSONFromIP(handler, "/api/configs", "127.0.0.1:12345", map[string]any{
		"key":   "APP_URL",
		"value": "https://example.test",
	})
	if save.Code != http.StatusOK {
		t.Fatalf("save expected 200, got %d: %s", save.Code, save.Body.String())
	}

	get := getJSON(handler, "/api/configs/APP_URL")
	if get.Code != http.StatusOK {
		t.Fatalf("get expected 200, got %d: %s", get.Code, get.Body.String())
	}
	var getResp map[string]any
	if err := json.Unmarshal(get.Body.Bytes(), &getResp); err != nil {
		t.Fatalf("decode get: %v", err)
	}
	if getResp["ref"] != "VE:LOCAL:APP_URL" {
		t.Fatalf("ref = %v, want VE:LOCAL:APP_URL", getResp["ref"])
	}

	list := getJSON(handler, "/api/configs")
	if list.Code != http.StatusOK {
		t.Fatalf("list expected 200, got %d: %s", list.Code, list.Body.String())
	}
	var listResp struct {
		Count int `json:"count"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &listResp); err != nil {
		t.Fatalf("decode list: %v", err)
	}
	if listResp.Count != 1 {
		t.Fatalf("count = %d, want 1", listResp.Count)
	}

	delReq := httptest.NewRequest(http.MethodDelete, "/api/configs/APP_URL", nil)
	delReq.RemoteAddr = "127.0.0.1:12345"
	del := httptest.NewRecorder()
	handler.ServeHTTP(del, delReq)
	if del.Code != http.StatusOK {
		t.Fatalf("delete expected 200, got %d: %s", del.Code, del.Body.String())
	}
}

func TestStatusEndpointIncludesLocalConfigCapability(t *testing.T) {
	srv, handler := setupTestServer(t)
	if err := srv.db.SaveConfig("DOMAIN", "example.test"); err != nil {
		t.Fatalf("seed config: %v", err)
	}

	w := getJSON(handler, "/api/status")
	if w.Code != http.StatusOK {
		t.Fatalf("status expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		SupportedFeatures []string `json:"supported_features"`
		ConfigsCount      int      `json:"configs_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode status: %v", err)
	}
	if resp.ConfigsCount != 1 {
		t.Fatalf("configs_count = %d, want 1", resp.ConfigsCount)
	}
	found := false
	for _, feature := range resp.SupportedFeatures {
		if feature == "configs" {
			found = true
			break
		}
	}
	if !found {
		t.Fatalf("supported_features missing configs: %#v", resp.SupportedFeatures)
	}
}
