package api

import (
	_ "embed"
	"net/http"
	"os"
	"path/filepath"
	"strings"
)

//go:embed admin_vue_preview.html
var adminVuePreviewHTML string

func (s *Server) handleAdminVuePreview(w http.ResponseWriter, r *http.Request) {
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
