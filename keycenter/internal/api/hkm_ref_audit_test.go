package api

import (
	"slices"
	"testing"
)

type auditSeed struct {
	canonical string
	version   int
	status    string
	owner     string
}

type auditIssueExpect struct {
	reason string
	key    string
	refs   []string
}

func TestTrackedRefAuditReasonMatrix(t *testing.T) {
	tests := []struct {
		name           string
		seeds          []auditSeed
		expectedBlocked []string
		expectedStale  []auditIssueExpect
		expectedCounts map[string]int
	}{
		{
			name: "blocked refs stay in blocked top-level class",
			seeds: []auditSeed{
				{canonical: "VK:LOCAL:deadbeef", version: 1, status: "block", owner: "agent-a"},
			},
			expectedBlocked: []string{"VK:LOCAL:deadbeef"},
			expectedCounts: map[string]int{
				"total_refs": 1,
				"blocked":    1,
				"stale":      0,
			},
		},
		{
			name: "duplicate ref id and missing agent become separate stale reasons in order",
			seeds: []auditSeed{
				{canonical: "VK:TEMP:deadbeef", version: 4, status: "temp", owner: "agent-a"},
				{canonical: "VK:LOCAL:deadbeef", version: 4, status: "active", owner: "agent-a"},
				{canonical: "VE:EXTERNAL:APP_URL", version: 4, status: "active", owner: "missing-owner"},
			},
			expectedStale: []auditIssueExpect{
				{
					reason: trackedRefAuditReasonDuplicateRefID,
					key:    "{AGENT_A}|VK|deadbeef",
					refs:   []string{"VK:LOCAL:deadbeef", "VK:TEMP:deadbeef"},
				},
				{
					reason: trackedRefAuditReasonMissingAgent,
					key:    "missing-owner",
					refs:   []string{"VE:EXTERNAL:APP_URL"},
				},
			},
			expectedCounts: map[string]int{
				"total_refs": 3,
				"blocked":    0,
				"stale":      2,
			},
		},
		{
			name: "missing owner and agent mismatch are distinct reasons",
			seeds: []auditSeed{
				{canonical: "VK:LOCAL:ownerless", version: 2, status: "active", owner: ""},
				{canonical: "VK:LOCAL:sharedref", version: 3, status: "active", owner: "agent-a"},
				{canonical: "VK:TEMP:sharedref", version: 2, status: "temp", owner: "agent-b"},
			},
			expectedStale: []auditIssueExpect{
				{
					reason: "agent_mismatch",
					key:    "VK|sharedref",
					refs:   []string{"VK:LOCAL:sharedref", "VK:TEMP:sharedref"},
				},
				{
					reason: "missing_owner",
					key:    "VK:LOCAL:ownerless",
					refs:   []string{"VK:LOCAL:ownerless"},
				},
			},
			expectedCounts: map[string]int{
				"total_refs": 3,
				"blocked":    0,
				"stale":      2,
			},
		},
		{
			name: "blocked and stale coexist without creating new top-level class",
			seeds: []auditSeed{
				{canonical: "VK:LOCAL:block0001", version: 5, status: "block", owner: "agent-a"},
				{canonical: "VK:TEMP:deadbeef", version: 5, status: "temp", owner: "agent-a"},
				{canonical: "VK:LOCAL:deadbeef", version: 5, status: "active", owner: "agent-a"},
				{canonical: "VE:EXTERNAL:APP_URL", version: 5, status: "active", owner: "missing-owner"},
			},
			expectedBlocked: []string{"VK:LOCAL:block0001"},
			expectedStale: []auditIssueExpect{
				{
					reason: trackedRefAuditReasonDuplicateRefID,
					key:    "{AGENT_A}|VK|deadbeef",
					refs:   []string{"VK:LOCAL:deadbeef", "VK:TEMP:deadbeef"},
				},
				{
					reason: trackedRefAuditReasonMissingAgent,
					key:    "missing-owner",
					refs:   []string{"VE:EXTERNAL:APP_URL"},
				},
			},
			expectedCounts: map[string]int{
				"total_refs": 4,
				"blocked":    1,
				"stale":      2,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv, handler := setupHKMServer(t)
			hb := postJSON(handler, "/api/agents/heartbeat", map[string]any{
				"node_id":     "node-audit",
				"label":       "audit-agent",
				"vault_hash":  "vh-audit",
				"vault_name":  "audit-vault",
				"key_version": 1,
				"ip":          "10.0.1.10",
				"port":        10180,
			})
			var resp struct {
				VaultRuntimeHash string `json:"vault_runtime_hash"`
			}
			decodeJSON(t, hb, &resp)

			hbB := postJSON(handler, "/api/agents/heartbeat", map[string]any{
				"node_id":     "node-audit-b",
				"label":       "audit-agent-b",
				"vault_hash":  "vh-audit-b",
				"vault_name":  "audit-vault-b",
				"key_version": 1,
				"ip":          "10.0.1.11",
				"port":        10181,
			})
			var respB struct {
				VaultRuntimeHash string `json:"vault_runtime_hash"`
			}
			decodeJSON(t, hbB, &respB)

			for _, seed := range tt.seeds {
				owner := seed.owner
				switch owner {
				case "agent-a":
					owner = resp.VaultRuntimeHash
				case "agent-b":
					owner = respB.VaultRuntimeHash
				}
				if err := srv.upsertTrackedRef(seed.canonical, seed.version, seed.status, owner); err != nil {
					t.Fatalf("seed tracked ref %q: %v", seed.canonical, err)
				}
			}

			w := getJSON(handler, "/api/tracked-refs/audit")
			if w.Code != 200 {
				t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
			}

			var audit struct {
				Blocked []trackedRefAuditEntry `json:"blocked"`
				Stale   []trackedRefAuditIssue `json:"stale"`
				Counts  map[string]int         `json:"counts"`
			}
			decodeJSON(t, w, &audit)

			if len(audit.Blocked) != len(tt.expectedBlocked) {
				t.Fatalf("blocked len = %d, want %d (%+v)", len(audit.Blocked), len(tt.expectedBlocked), audit.Blocked)
			}
			for i, want := range tt.expectedBlocked {
				if audit.Blocked[i].RefCanonical != want {
					t.Fatalf("blocked[%d] = %q, want %q", i, audit.Blocked[i].RefCanonical, want)
				}
			}

			if len(audit.Stale) != len(tt.expectedStale) {
				t.Fatalf("stale len = %d, want %d (%+v)", len(audit.Stale), len(tt.expectedStale), audit.Stale)
			}
			for i, want := range tt.expectedStale {
				got := audit.Stale[i]
				if got.Reason != want.reason {
					t.Fatalf("stale[%d].reason = %q, want %q", i, got.Reason, want.reason)
				}
				wantKey := want.key
				if wantKey == "{AGENT_A}|VK|deadbeef" {
					wantKey = resp.VaultRuntimeHash + "|VK|deadbeef"
				}
				if got.Key != wantKey {
					t.Fatalf("stale[%d].key = %q, want %q", i, got.Key, wantKey)
				}
				if len(got.Refs) != len(want.refs) {
					t.Fatalf("stale[%d].refs len = %d, want %d (%+v)", i, len(got.Refs), len(want.refs), got.Refs)
				}
				gotRefs := make([]string, 0, len(got.Refs))
				for _, ref := range got.Refs {
					gotRefs = append(gotRefs, ref.RefCanonical)
				}
				wantRefs := append([]string(nil), want.refs...)
				slices.Sort(gotRefs)
				slices.Sort(wantRefs)
				for j := range want.refs {
					if gotRefs[j] != wantRefs[j] {
						t.Fatalf("stale[%d].refs[%d] = %q, want %q", i, j, gotRefs[j], wantRefs[j])
					}
				}
			}

			for key, want := range tt.expectedCounts {
				if audit.Counts[key] != want {
					t.Fatalf("counts[%q] = %d, want %d", key, audit.Counts[key], want)
				}
			}
		})
	}
}
