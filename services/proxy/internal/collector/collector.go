package collector

import (
	"context"

	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/config"
	"gitlab.ranode.net/veilkey/veilkey-proxy/internal/events"
)

type Collector interface {
	Preflight() error
	Observe(context.Context, func(events.Event)) error
	Close() error
}

func New(cfg config.Config) (Collector, error) {
	return newLinuxCollector(cfg)
}
