package api

import "time"

func (s *Server) advancePendingRotationsBestEffort() {
	_, _ = s.db.AdvancePendingRotations(time.Now().UTC())
}
