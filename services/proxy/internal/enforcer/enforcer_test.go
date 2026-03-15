package enforcer

import (
	"testing"

	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/config"
	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/events"
)

func TestApplyNoopWhenDisabled(t *testing.T) {
	e := New(config.Config{})
	ev := e.Apply(events.Event{
		Kind:       events.KindExecve,
		PID:        123,
		Suspicious: true,
	})
	if ev.EnforcementAction != "" {
		t.Fatalf("expected no enforcement action, got %q", ev.EnforcementAction)
	}
}

func TestApplyNoopForCleanEvent(t *testing.T) {
	e := New(config.Config{EnforceKill: true})
	ev := e.Apply(events.Event{
		Kind:       events.KindExecve,
		PID:        123,
		Suspicious: false,
	})
	if ev.EnforcementAction != "" {
		t.Fatalf("expected no enforcement action, got %q", ev.EnforcementAction)
	}
}
