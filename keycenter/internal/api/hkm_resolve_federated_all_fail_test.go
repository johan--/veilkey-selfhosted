package api

import (
	"net/http"
	"testing"
	"veilkey-keycenter/internal/crypto"
	"veilkey-keycenter/internal/db"
)

func TestHKM_ResolveFederatedAllChildrenFail(t *testing.T) {
	srv, handler := setupHKMServer(t)

	mockChild := newMock404Server()
	defer mockChild.Close()

	srv.db.RegisterChild(&db.Child{
		NodeID: crypto.GenerateUUID(), Label: "404-child", URL: mockChild.URL, Version: 1,
	})

	w := getJSON(handler, "/api/resolve/missing_everywhere")
	if w.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", w.Code)
	}
}
