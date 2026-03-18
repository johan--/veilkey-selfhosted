package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHKM_StatusAndNodeInfoParity(t *testing.T) {
	_, handler := setupHKMServer(t)

	statusW := getJSON(handler, "/api/status")
	if statusW.Code != http.StatusOK {
		t.Fatalf("status expected 200, got %d: %s", statusW.Code, statusW.Body.String())
	}
	nodeInfoW := getJSON(handler, "/api/node-info")
	if nodeInfoW.Code != http.StatusOK {
		t.Fatalf("node-info expected 200, got %d: %s", nodeInfoW.Code, nodeInfoW.Body.String())
	}

	var statusResp struct {
		Mode             string `json:"mode"`
		NodeID           string `json:"node_id"`
		Version          int    `json:"version"`
		ChildrenCount    int    `json:"children_count"`
		TrackedRefsCount int    `json:"tracked_refs_count"`
		Locked           bool   `json:"locked"`
	}
	if err := json.Unmarshal(statusW.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("decode status: %v", err)
	}

	var nodeInfoResp struct {
		Mode             string `json:"mode"`
		NodeID           string `json:"node_id"`
		Version          int    `json:"version"`
		ChildrenCount    int    `json:"children_count"`
		TrackedRefsCount int    `json:"tracked_refs_count"`
	}
	if err := json.Unmarshal(nodeInfoW.Body.Bytes(), &nodeInfoResp); err != nil {
		t.Fatalf("decode node-info: %v", err)
	}

	if statusResp.Mode != "hkm" || nodeInfoResp.Mode != "hkm" {
		t.Fatalf("mode mismatch: status=%q node-info=%q", statusResp.Mode, nodeInfoResp.Mode)
	}
	if statusResp.NodeID != nodeInfoResp.NodeID {
		t.Fatalf("node_id mismatch: status=%q node-info=%q", statusResp.NodeID, nodeInfoResp.NodeID)
	}
	if statusResp.Version != nodeInfoResp.Version {
		t.Fatalf("version mismatch: status=%d node-info=%d", statusResp.Version, nodeInfoResp.Version)
	}
	if statusResp.ChildrenCount != nodeInfoResp.ChildrenCount {
		t.Fatalf("children_count mismatch: status=%d node-info=%d", statusResp.ChildrenCount, nodeInfoResp.ChildrenCount)
	}
	if statusResp.TrackedRefsCount != nodeInfoResp.TrackedRefsCount {
		t.Fatalf("tracked_refs_count mismatch: status=%d node-info=%d", statusResp.TrackedRefsCount, nodeInfoResp.TrackedRefsCount)
	}
	if statusResp.Locked {
		t.Fatal("status locked should be false in unlocked HKM test server")
	}
}

func TestHKM_StatusAndNodeInfoParityWithTrackedRefAndChildCounts(t *testing.T) {
	srv, handler := setupHKMServer(t)
	seedCentralRuntimeCounts(t, srv, handler)

	statusW := getJSON(handler, "/api/status")
	if statusW.Code != http.StatusOK {
		t.Fatalf("status expected 200, got %d: %s", statusW.Code, statusW.Body.String())
	}
	nodeInfoW := getJSON(handler, "/api/node-info")
	if nodeInfoW.Code != http.StatusOK {
		t.Fatalf("node-info expected 200, got %d: %s", nodeInfoW.Code, nodeInfoW.Body.String())
	}

	var statusResp struct {
		Mode             string `json:"mode"`
		NodeID           string `json:"node_id"`
		VaultNodeUUID    string `json:"vault_node_uuid"`
		Version          int    `json:"version"`
		ChildrenCount    int    `json:"children_count"`
		TrackedRefsCount int    `json:"tracked_refs_count"`
		Locked           bool   `json:"locked"`
	}
	if err := json.Unmarshal(statusW.Body.Bytes(), &statusResp); err != nil {
		t.Fatalf("decode status: %v", err)
	}

	var nodeInfoResp struct {
		Mode             string `json:"mode"`
		NodeID           string `json:"node_id"`
		VaultNodeUUID    string `json:"vault_node_uuid"`
		Version          int    `json:"version"`
		ChildrenCount    int    `json:"children_count"`
		TrackedRefsCount int    `json:"tracked_refs_count"`
	}
	if err := json.Unmarshal(nodeInfoW.Body.Bytes(), &nodeInfoResp); err != nil {
		t.Fatalf("decode node-info: %v", err)
	}

	if statusResp.Mode != nodeInfoResp.Mode {
		t.Fatalf("mode mismatch: status=%q node-info=%q", statusResp.Mode, nodeInfoResp.Mode)
	}
	if statusResp.NodeID != nodeInfoResp.NodeID {
		t.Fatalf("node_id mismatch: status=%q node-info=%q", statusResp.NodeID, nodeInfoResp.NodeID)
	}
	if statusResp.VaultNodeUUID != nodeInfoResp.VaultNodeUUID {
		t.Fatalf("vault_node_uuid mismatch: status=%q node-info=%q", statusResp.VaultNodeUUID, nodeInfoResp.VaultNodeUUID)
	}
	if statusResp.Version != nodeInfoResp.Version {
		t.Fatalf("version mismatch: status=%d node-info=%d", statusResp.Version, nodeInfoResp.Version)
	}
	if statusResp.ChildrenCount != 1 || nodeInfoResp.ChildrenCount != 1 {
		t.Fatalf("children_count mismatch: status=%d node-info=%d", statusResp.ChildrenCount, nodeInfoResp.ChildrenCount)
	}
	if statusResp.TrackedRefsCount != 1 || nodeInfoResp.TrackedRefsCount != 1 {
		t.Fatalf("tracked_refs_count mismatch: status=%d node-info=%d", statusResp.TrackedRefsCount, nodeInfoResp.TrackedRefsCount)
	}
	if statusResp.Locked {
		t.Fatal("status locked should be false in unlocked HKM test server")
	}
}
