package api

import (
	"net/http"
	"testing"
)

func TestHKM_Resolve_PathTraversal(t *testing.T) {
	_, handler := setupHKMServer(t)

	w := getJSON(handler, "/api/resolve/../../etc/passwd")
	if w.Code == http.StatusOK {
		t.Error("path traversal should not succeed")
	}
}
