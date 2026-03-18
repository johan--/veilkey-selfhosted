package api

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func decodeJSON(t *testing.T, w *httptest.ResponseRecorder, dst interface{}) {
	t.Helper()
	if err := json.Unmarshal(w.Body.Bytes(), dst); err != nil {
		t.Fatalf("decodeJSON: %v; body=%s", err, w.Body.String())
	}
}
