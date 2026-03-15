package api

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHKMStatusTrackedRefsContractFollowsLifecycle(t *testing.T) {
	srv, handler := setupHKMServer(t)

	hb := postJSON(handler, "/api/agents/heartbeat", map[string]any{
		"vault_node_uuid": "node-status-contract",
		"label":           "status-contract-agent",
		"vault_hash":      "vh-status-contract",
		"vault_name":      "status-contract-vault",
		"key_version":     7,
		"ip":              "10.0.1.21",
		"port":            10180,
	})
	if hb.Code != http.StatusOK {
		t.Fatalf("heartbeat expected 200, got %d: %s", hb.Code, hb.Body.String())
	}

	assertCounts := func(want int) {
		t.Helper()
		var statusResp struct {
			TrackedRefsCount int    `json:"tracked_refs_count"`
			NodeID           string `json:"node_id"`
			VaultNodeUUID    string `json:"vault_node_uuid"`
		}
		statusW := getJSON(handler, "/api/status")
		if statusW.Code != http.StatusOK {
			t.Fatalf("status expected 200, got %d: %s", statusW.Code, statusW.Body.String())
		}
		if err := json.Unmarshal(statusW.Body.Bytes(), &statusResp); err != nil {
			t.Fatalf("decode status: %v", err)
		}

		var nodeInfoResp struct {
			TrackedRefsCount int    `json:"tracked_refs_count"`
			NodeID           string `json:"node_id"`
			VaultNodeUUID    string `json:"vault_node_uuid"`
		}
		nodeInfoW := getJSON(handler, "/api/node-info")
		if nodeInfoW.Code != http.StatusOK {
			t.Fatalf("node-info expected 200, got %d: %s", nodeInfoW.Code, nodeInfoW.Body.String())
		}
		if err := json.Unmarshal(nodeInfoW.Body.Bytes(), &nodeInfoResp); err != nil {
			t.Fatalf("decode node-info: %v", err)
		}

		if statusResp.TrackedRefsCount != want {
			t.Fatalf("status tracked_refs_count = %d, want %d", statusResp.TrackedRefsCount, want)
		}
		if nodeInfoResp.TrackedRefsCount != want {
			t.Fatalf("node-info tracked_refs_count = %d, want %d", nodeInfoResp.TrackedRefsCount, want)
		}
		if statusResp.NodeID != nodeInfoResp.NodeID {
			t.Fatalf("node_id mismatch: status=%q node-info=%q", statusResp.NodeID, nodeInfoResp.NodeID)
		}
		if statusResp.VaultNodeUUID != nodeInfoResp.VaultNodeUUID {
			t.Fatalf("vault_node_uuid mismatch: status=%q node-info=%q", statusResp.VaultNodeUUID, nodeInfoResp.VaultNodeUUID)
		}
	}

	assertCounts(0)

	saveTemp := postJSON(handler, "/api/tracked-refs/sync", map[string]any{
		"vault_node_uuid": "node-status-contract",
		"ref":             "VK:TEMP:status001",
		"version":         7,
		"status":          "temp",
	})
	if saveTemp.Code != http.StatusOK {
		t.Fatalf("save temp expected 200, got %d: %s", saveTemp.Code, saveTemp.Body.String())
	}
	assertCounts(1)

	activateLocal := postJSON(handler, "/api/tracked-refs/sync", map[string]any{
		"vault_node_uuid": "node-status-contract",
		"ref":             "VK:LOCAL:status001",
		"previous_ref":    "VK:TEMP:status001",
		"version":         7,
		"status":          "active",
	})
	if activateLocal.Code != http.StatusOK {
		t.Fatalf("activate local expected 200, got %d: %s", activateLocal.Code, activateLocal.Body.String())
	}
	assertCounts(1)

	saveConfig := postJSON(handler, "/api/tracked-refs/sync", map[string]any{
		"vault_node_uuid": "node-status-contract",
		"ref":             "VE:LOCAL:APP_URL",
		"version":         7,
		"status":          "active",
	})
	if saveConfig.Code != http.StatusOK {
		t.Fatalf("save config expected 200, got %d: %s", saveConfig.Code, saveConfig.Body.String())
	}
	assertCounts(2)

	if err := srv.deleteTrackedRef("VE:LOCAL:APP_URL"); err != nil {
		t.Fatalf("delete config tracked ref: %v", err)
	}
	assertCounts(1)
}
