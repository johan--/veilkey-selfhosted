package api

import (
	"testing"
	"time"
)

func TestSetTimeouts(t *testing.T) {
	srv, _ := setupHKMServer(t)

	custom := Timeouts{CascadeResolve: 10 * time.Second, ParentForward: 7 * time.Second, Deploy: 60 * time.Second}
	srv.SetTimeouts(custom)

	if srv.timeouts.CascadeResolve != 10*time.Second {
		t.Errorf("CascadeResolve = %v, want 10s", srv.timeouts.CascadeResolve)
	}
	if srv.timeouts.ParentForward != 7*time.Second {
		t.Errorf("ParentForward = %v, want 7s", srv.timeouts.ParentForward)
	}
	if srv.timeouts.Deploy != 60*time.Second {
		t.Errorf("Deploy = %v, want 60s", srv.timeouts.Deploy)
	}
}
