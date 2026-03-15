package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

func (s *Server) handleOperatorShellEntry(w http.ResponseWriter, r *http.Request) {
	if s.IsLocked() {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(lockedLandingHTML))
		return
	}
	if complete, session := s.installGateState(); !complete {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		renderInstallGate(w, session)
		return
	}
	s.handleDashboard(w, r)
}

func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if devDir := strings.TrimSpace(os.Getenv("VEILKEY_UI_DEV_DIR")); devDir != "" {
		path := filepath.Join(devDir, "admin_vue_preview.html")
		if body, err := os.ReadFile(path); err == nil {
			_, _ = w.Write(body)
			return
		}
	}
	_, _ = w.Write([]byte(adminVuePreviewHTML))
}
