package api

import (
	"net/http"
	"net/http/httptest"
)

func newMock404Server() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.NotFound(w, r)
	}))
}
