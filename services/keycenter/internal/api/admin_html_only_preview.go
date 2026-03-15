package api

import (
	_ "embed"
	"net/http"
)

//go:embed admin_html_only_preview.html
var adminHTMLOneShotPreviewHTML string

func (s *Server) handleAdminHTMLOneShotPreview(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(adminHTMLOneShotPreviewHTML))
}
