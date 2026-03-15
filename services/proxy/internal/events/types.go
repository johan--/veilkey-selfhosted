package events

import "time"

type Kind string

const (
	KindExecve  Kind = "execve"
	KindConnect Kind = "connect"
)

type Event struct {
	Time       time.Time `json:"time"`
	Kind       Kind      `json:"kind"`
	PID        uint32    `json:"pid"`
	PPID       uint32    `json:"ppid"`
	UID        uint32    `json:"uid"`
	Comm       string    `json:"comm"`
	CgroupPath string    `json:"cgroup_path,omitempty"`
	TargetAddr string    `json:"target_addr,omitempty"`
	Argv       []string  `json:"argv,omitempty"`
	Truncated  bool      `json:"truncated,omitempty"`
	Suspicious bool      `json:"suspicious,omitempty"`
	Matches    []string  `json:"matches,omitempty"`
	EnforcementAction string `json:"enforcement_action,omitempty"`
	EnforcementStatus string `json:"enforcement_status,omitempty"`
	EnforcementError  string `json:"enforcement_error,omitempty"`
}
