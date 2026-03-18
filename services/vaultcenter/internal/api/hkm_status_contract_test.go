package api

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type centralRuntimeInfoResponse struct {
	Mode             string `json:"mode"`
	NodeID           string `json:"node_id"`
	VaultNodeUUID    string `json:"vault_node_uuid"`
	Version          int    `json:"version"`
	ChildrenCount    int    `json:"children_count"`
	TrackedRefsCount int    `json:"tracked_refs_count"`
	Locked           bool   `json:"locked"`
}

func decodeCentralRuntimeInfo(t *testing.T, w *httptest.ResponseRecorder) centralRuntimeInfoResponse {
	t.Helper()

	var resp centralRuntimeInfoResponse
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode runtime info: %v", err)
	}
	return resp
}

func seedCentralRuntimeCounts(t *testing.T, srv *Server, handler http.Handler) {
	t.Helper()

	if err := srv.upsertTrackedRef("VK:LOCAL:status-contract-01", 1, "active", "runtime-owner"); err != nil {
		t.Fatalf("seed tracked ref: %v", err)
	}

	w := postJSON(handler, "/api/register", map[string]string{
		"node_id": "child-status-contract-01",
		"label":   "status-contract-child",
		"url":     "http://child-status-contract-01:10180",
	})
	if w.Code != http.StatusOK {
		t.Fatalf("register child: expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestStatusEndpointReflectsCentralTrackedRefContract(t *testing.T) {
	srv, handler := setupHKMServer(t)
	seedCentralRuntimeCounts(t, srv, handler)

	w := getJSON(handler, "/api/status")
	if w.Code != http.StatusOK {
		t.Fatalf("status expected 200, got %d: %s", w.Code, w.Body.String())
	}

	resp := decodeCentralRuntimeInfo(t, w)
	if resp.Mode != "hkm" {
		t.Fatalf("mode = %q, want hkm", resp.Mode)
	}
	if resp.NodeID == "" || resp.VaultNodeUUID == "" {
		t.Fatalf("node identifiers should not be empty: %+v", resp)
	}
	if resp.NodeID != resp.VaultNodeUUID {
		t.Fatalf("vault_node_uuid = %q, want %q", resp.VaultNodeUUID, resp.NodeID)
	}
	if resp.Version != 1 {
		t.Fatalf("version = %d, want 1", resp.Version)
	}
	if resp.ChildrenCount != 1 {
		t.Fatalf("children_count = %d, want 1", resp.ChildrenCount)
	}
	if resp.TrackedRefsCount != 1 {
		t.Fatalf("tracked_refs_count = %d, want 1", resp.TrackedRefsCount)
	}
	if resp.Locked {
		t.Fatal("locked should be false")
	}
}

func TestNodeInfoReflectsCentralTrackedRefContract(t *testing.T) {
	srv, handler := setupHKMServer(t)
	seedCentralRuntimeCounts(t, srv, handler)

	w := getJSON(handler, "/api/node-info")
	if w.Code != http.StatusOK {
		t.Fatalf("node-info expected 200, got %d: %s", w.Code, w.Body.String())
	}

	var resp struct {
		Mode             string `json:"mode"`
		NodeID           string `json:"node_id"`
		VaultNodeUUID    string `json:"vault_node_uuid"`
		Version          int    `json:"version"`
		ChildrenCount    int    `json:"children_count"`
		TrackedRefsCount int    `json:"tracked_refs_count"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode node-info: %v", err)
	}
	if resp.Mode != "hkm" {
		t.Fatalf("mode = %q, want hkm", resp.Mode)
	}
	if resp.NodeID == "" || resp.VaultNodeUUID == "" {
		t.Fatalf("node identifiers should not be empty: %+v", resp)
	}
	if resp.NodeID != resp.VaultNodeUUID {
		t.Fatalf("vault_node_uuid = %q, want %q", resp.VaultNodeUUID, resp.NodeID)
	}
	if resp.Version != 1 {
		t.Fatalf("version = %d, want 1", resp.Version)
	}
	if resp.ChildrenCount != 1 {
		t.Fatalf("children_count = %d, want 1", resp.ChildrenCount)
	}
	if resp.TrackedRefsCount != 1 {
		t.Fatalf("tracked_refs_count = %d, want 1", resp.TrackedRefsCount)
	}
}

func TestStatusAndNodeInfoShareLockedGate(t *testing.T) {
	_, handler := setupLockedServer(t)

	statusW := getJSON(handler, "/api/status")
	if statusW.Code != http.StatusServiceUnavailable {
		t.Fatalf("status expected 503, got %d: %s", statusW.Code, statusW.Body.String())
	}

	nodeInfoW := getJSON(handler, "/api/node-info")
	if nodeInfoW.Code != http.StatusServiceUnavailable {
		t.Fatalf("node-info expected 503, got %d: %s", nodeInfoW.Code, nodeInfoW.Body.String())
	}

	var statusErr map[string]string
	if err := json.Unmarshal(statusW.Body.Bytes(), &statusErr); err != nil {
		t.Fatalf("decode status error: %v", err)
	}
	var nodeInfoErr map[string]string
	if err := json.Unmarshal(nodeInfoW.Body.Bytes(), &nodeInfoErr); err != nil {
		t.Fatalf("decode node-info error: %v", err)
	}
	if statusErr["error"] != nodeInfoErr["error"] {
		t.Fatalf("locked error mismatch: status=%q node-info=%q", statusErr["error"], nodeInfoErr["error"])
	}
}
