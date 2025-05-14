package lb

import (
	"cloud-test/internal/backends"
	"context"
	"errors"
)

var _ Balancer = (*RoundRobin)(nil)

type Balancer interface {
	Next() (backends.Backend, error)
}

type RoundRobin struct {
	manager *backends.Registry
	idx     int
}

func NewRoundRobin(m *backends.Registry, ctx context.Context) (*RoundRobin, error) {
	return &RoundRobin{manager: m}, nil
}

func (r *RoundRobin) Next() (backends.Backend, error) {
	aliveBackends := r.manager.AliveBackends()
	if len(aliveBackends) == 0 {
		return backends.Backend{}, errors.New("no alive backends available")
	}

	selectedBackend := aliveBackends[r.idx%len(aliveBackends)]
	r.idx = (r.idx + 1) % len(aliveBackends)

	return selectedBackend, nil
}
