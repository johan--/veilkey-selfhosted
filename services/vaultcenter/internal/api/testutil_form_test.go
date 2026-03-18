package api

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
)

func postForm(handler http.Handler, path string, values map[string]string) *httptest.ResponseRecorder {
	form := url.Values{}
	for k, v := range values {
		form.Set(k, v)
	}
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}
