package api

import "net/http"

func (s *Server) handleAdminHTMLOneShotPreview(w http.ResponseWriter, r *http.Request) {
	s.handleAdminVuePreview(w, r)
}
