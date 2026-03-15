package api

import (
	"testing"
	"time"
)

func TestDefaultTimeouts(t *testing.T) {
	defaults := DefaultTimeouts()
	if defaults.CascadeResolve != 5*time.Second {
		t.Errorf("CascadeResolve = %v, want 5s", defaults.CascadeResolve)
	}
	if defaults.ParentForward != 3*time.Second {
		t.Errorf("ParentForward = %v, want 3s", defaults.ParentForward)
	}
	if defaults.Deploy != 30*time.Second {
		t.Errorf("Deploy = %v, want 30s", defaults.Deploy)
	}
}
