//go:build linux

package collector

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/cilium/ebpf/link"
	"github.com/cilium/ebpf/ringbuf"
	"github.com/cilium/ebpf/rlimit"

	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/config"
	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/events"
)

type linuxCollector struct {
	cfg           config.Config
	execveObjs    *execveProbeObjects
	execveTP      link.Link
	execveReader  *ringbuf.Reader
	connectObjs   *connectProbeObjects
	connectTP     link.Link
	connectReader *ringbuf.Reader
}

func newLinuxCollector(cfg config.Config) (Collector, error) {
	return &linuxCollector{cfg: cfg}, nil
}

func (c *linuxCollector) Preflight() error {
	if runtime.GOOS != "linux" {
		return errors.New("linux collector requires linux")
	}
	if os.Geteuid() != 0 {
		return errors.New("linux collector requires root")
	}
	if err := rlimit.RemoveMemlock(); err != nil {
		return fmt.Errorf("remove memlock rlimit: %w", err)
	}
	if err := requirePath("/sys/fs/bpf"); err != nil {
		return err
	}
	if c.cfg.TargetCgroup != "" {
		if strings.HasPrefix(c.cfg.TargetCgroup, "/sys/fs/cgroup/") {
			if err := requirePath(c.cfg.TargetCgroup); err != nil {
				return fmt.Errorf("target cgroup: %w", err)
			}
		}
	}
	return nil
}

func (c *linuxCollector) Observe(ctx context.Context, emit func(events.Event)) error {
	errCh := make(chan error, 2)

	go func() {
		errCh <- c.observeExec(ctx, emit)
	}()
	go func() {
		errCh <- c.observeConnect(ctx, emit)
	}()

	for i := 0; i < 2; i++ {
		err := <-errCh
		if err != nil {
			return err
		}
	}
	return nil
}

func (c *linuxCollector) Close() error {
	if c.execveReader != nil {
		_ = c.execveReader.Close()
	}
	if c.execveTP != nil {
		_ = c.execveTP.Close()
	}
	if c.execveObjs != nil {
		_ = c.execveObjs.Close()
	}
	if c.connectReader != nil {
		_ = c.connectReader.Close()
	}
	if c.connectTP != nil {
		_ = c.connectTP.Close()
	}
	if c.connectObjs != nil {
		_ = c.connectObjs.Close()
	}
	return nil
}

func requirePath(path string) error {
	if path == "" {
		return errors.New("path is empty")
	}
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("%s: %w", filepath.Clean(path), err)
	}
	return nil
}

func procCgroupPath(pid uint32) string {
	data, err := os.ReadFile(fmt.Sprintf("/proc/%d/cgroup", pid))
	if err != nil {
		return ""
	}
	lines := strings.Split(string(data), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		parts := strings.SplitN(line, ":", 3)
		if len(parts) == 3 {
			return parts[2]
		}
	}
	return ""
}

func normalizeCgroupMatch(value string) string {
	prefix := "/sys/fs/cgroup"
	if strings.HasPrefix(value, prefix) {
		out := strings.TrimPrefix(value, prefix)
		if out == "" {
			return "/"
		}
		return out
	}
	return value
}
