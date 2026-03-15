package integration_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
)

func postJSONFromIP(handler http.Handler, path, remoteAddr string, body interface{}) *httptest.ResponseRecorder {
	b, _ := json.Marshal(body)
	req := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	req.RemoteAddr = remoteAddr
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	return w
}
