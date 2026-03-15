package api

import "net/http"

func (s *Server) handleListConfigs(w http.ResponseWriter, r *http.Request) {
	configs, err := s.db.ListConfigs()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list configs")
		return
	}

	type configResp struct {
		Key    string `json:"key"`
		Value  string `json:"value"`
		Ref    string `json:"ref"`
		Scope  string `json:"scope"`
		Status string `json:"status"`
	}
	result := make([]configResp, 0, len(configs))
	for _, c := range configs {
		result = append(result, configResp{
			Key:    c.Key,
			Value:  c.Value,
			Ref:    ParsedRef{Family: RefFamilyVE, Scope: RefScope(c.Scope), ID: c.Key}.CanonicalString(),
			Scope:  c.Scope,
			Status: c.Status,
		})
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"configs": result,
		"count":   len(result),
	})
}
