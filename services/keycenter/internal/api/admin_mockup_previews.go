package api

import "net/http"

func (s *Server) handleAdminMockupDark(w http.ResponseWriter, r *http.Request) {
	s.handleAdminVuePreview(w, r)
}

func (s *Server) handleAdminMockupAmber(w http.ResponseWriter, r *http.Request) {
	s.handleAdminVuePreview(w, r)
}

func (s *Server) handleAdminMockupMono(w http.ResponseWriter, r *http.Request) {
	s.handleAdminVuePreview(w, r)
}
