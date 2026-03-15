package api

import (
	"net/http"
	"net/http/httptest"
)

func deleteJSON(handler http.Handler, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodDelete, path, nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}
