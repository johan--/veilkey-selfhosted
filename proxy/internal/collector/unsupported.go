//go:build !linux

package collector

import (
	"context"
	"errors"

	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/config"
	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/events"
)

type unsupportedCollector struct{}

func newLinuxCollector(_ config.Config) (Collector, error) {
	return &unsupportedCollector{}, nil
}

func (c *unsupportedCollector) Preflight() error {
	return errors.New("collector is only supported on linux")
}

func (c *unsupportedCollector) Observe(_ context.Context, _ func(events.Event)) error {
	return errors.New("collector is only supported on linux")
}

func (c *unsupportedCollector) Close() error {
	return nil
}
