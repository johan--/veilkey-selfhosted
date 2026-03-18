package integration_test

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestIntegration_RealHTTPServer(t *testing.T) {
	_, handler := setupTestServer(t)

	ts := httptest.NewServer(handler)
	defer ts.Close()
	client := ts.Client()

	doGet := func(path string) int {
		resp, err := client.Get(ts.URL + path)
		if err != nil {
			t.Fatalf("GET %s: %v", path, err)
		}
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		return resp.StatusCode
	}

	if code := doGet("/health"); code != http.StatusOK {
		t.Fatalf("health: expected 200, got %d", code)
	}

	if code := doGet("/api/status"); code != http.StatusOK {
		t.Fatalf("status: expected 200, got %d", code)
	}

	if code := doGet("/unknown-path"); code != http.StatusNotFound {
		t.Errorf("unknown: expected 404, got %d", code)
	}

	if code := doGet("/api/encrypt"); code != http.StatusNotFound {
		t.Errorf("removed encrypt: expected 404, got %d", code)
	}
}
