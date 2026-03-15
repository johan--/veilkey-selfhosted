package api

import (
	"net/http"
	"slices"
	"veilkey-keycenter/internal/db"
)

func (s *Server) handleRefPolicy(w http.ResponseWriter, r *http.Request) {
	policies := db.ListRefPolicies()
	resp := make([]map[string]any, 0, len(policies))
	for _, policy := range policies {
		scopes := make([]string, 0, len(policy.AllowedScopes))
		defaultStatuses := map[string]string{}
		for scope, status := range policy.AllowedScopes {
			scopes = append(scopes, scope)
			defaultStatuses[scope] = status
		}
		slices.Sort(scopes)
		resp = append(resp, map[string]any{
			"family":           policy.Family,
			"default_scope":    policy.DefaultScope,
			"allowed_scopes":   scopes,
			"default_statuses": defaultStatuses,
		})
	}
	s.respondJSON(w, http.StatusOK, map[string]any{
		"policies": resp,
		"count":    len(resp),
	})
}
