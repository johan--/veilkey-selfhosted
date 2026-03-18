package api

import (
	"net/http"
	"testing"
)

func TestHKM_Resolve_EmptyRef(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/resolve/")
	if w.Code == http.StatusOK {
		t.Error("empty ref should not succeed")
	}
}
