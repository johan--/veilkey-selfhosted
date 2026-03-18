package api

import "net/http"

func (s *Server) handleConfigsSummary(w http.ResponseWriter, r *http.Request) {
	agents, err := s.db.ListAgents()
	if err != nil {
		s.respondError(w, http.StatusInternalServerError, "failed to list agents")
		return
	}

	totalConfigs := 0
	agentsWithConfigs := 0
	for _, a := range agents {
		totalConfigs += a.ConfigsCount
		if a.ConfigsCount > 0 {
			agentsWithConfigs++
		}
	}

	s.respondJSON(w, http.StatusOK, map[string]interface{}{
		"total_configs":      totalConfigs,
		"total_agents":       len(agents),
		"agents_with_configs": agentsWithConfigs,
	})
}
