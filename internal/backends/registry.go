package backends

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

type Registry struct {
	backends    []Backend
	healthCheck *healthCheck
	alive       map[url.URL]bool
	mutex       sync.RWMutex
	client      *http.Client
}

func NewRegistry(ctx context.Context, c Configuration, httpClient *http.Client) (*Registry, error) {
	backends, err := parseBackendsFromConfiguration(c)
	if err != nil {
		return nil, err
	}

	healthCheck, err := parseHealthCheck(c)
	if err != nil {
		return nil, fmt.Errorf("failed to parse health check: %w", err)
	}

	alive := make(map[url.URL]bool, len(backends))
	for _, b := range backends {
		isAlive, err := healthCheck.checkBackend(ctx, b, httpClient)
		if err != nil {
			return nil, fmt.Errorf("failed to check backend %s: %w", b.Name, err)
		}
		alive[b.Url] = isAlive
	}

	m := &Registry{backends: backends, alive: alive, client: httpClient, healthCheck: &healthCheck}

	go m.runHealthCheck(ctx)

	return m, nil
}

func (m *Registry) AliveBackends() []Backend {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	aliveBackends := make([]Backend, 0, len(m.backends))
	for _, backend := range m.backends {
		if m.alive[backend.Url] {
			aliveBackends = append(aliveBackends, backend)
		}
	}
	return aliveBackends
}

func (m *Registry) runHealthCheck(ctx context.Context) {
	ticker := time.NewTicker(m.healthCheck.Interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			for _, backend := range m.backends {
				alive, _ := m.healthCheck.checkBackend(ctx, backend, m.client)

				m.mutex.Lock()
				m.alive[backend.Url] = alive
				m.mutex.Unlock()
			}
		}
	}
}
