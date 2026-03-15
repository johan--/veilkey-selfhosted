package api

import (
	_ "embed"
	"net/http"
)

//go:embed admin_mockup_dark_enterprise.html
var adminMockupDarkEnterpriseHTML string

//go:embed admin_mockup_industrial_amber.html
var adminMockupIndustrialAmberHTML string

//go:embed admin_mockup_monochrome_editorial.html
var adminMockupMonochromeEditorialHTML string

func serveEmbeddedHTML(w http.ResponseWriter, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(html))
}

func (s *Server) handleAdminMockupDark(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedHTML(w, adminMockupDarkEnterpriseHTML)
}

func (s *Server) handleAdminMockupAmber(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedHTML(w, adminMockupIndustrialAmberHTML)
}

func (s *Server) handleAdminMockupMono(w http.ResponseWriter, r *http.Request) {
	serveEmbeddedHTML(w, adminMockupMonochromeEditorialHTML)
}
