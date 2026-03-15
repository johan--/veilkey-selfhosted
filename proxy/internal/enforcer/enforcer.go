package enforcer

import (
	"fmt"
	"syscall"

	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/config"
	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/events"
)

type Enforcer struct {
	cfg config.Config
}

func New(cfg config.Config) Enforcer {
	return Enforcer{cfg: cfg}
}

func (e Enforcer) Apply(ev events.Event) events.Event {
	if !e.cfg.EnforceKill {
		return ev
	}
	if !ev.Suspicious || ev.Kind != events.KindExecve || ev.PID == 0 {
		return ev
	}
	err := syscall.Kill(int(ev.PID), syscall.SIGKILL)
	ev.EnforcementAction = "kill"
	if err != nil {
		ev.EnforcementStatus = "error"
		ev.EnforcementError = err.Error()
		return ev
	}
	ev.EnforcementStatus = "applied"
	return ev
}

func Describe(ev events.Event) string {
	if ev.EnforcementAction == "" {
		return ""
	}
	if ev.EnforcementError != "" {
		return fmt.Sprintf("%s:%s", ev.EnforcementAction, ev.EnforcementError)
	}
	return fmt.Sprintf("%s:%s", ev.EnforcementAction, ev.EnforcementStatus)
}
