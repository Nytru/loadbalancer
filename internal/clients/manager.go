package clients

import (
	"cloud-test/internal/datebase"
	"context"
	"log/slog"
	"sync"
	"time"
)

type ClientLimitManager struct {
	mu             sync.RWMutex
	clientSettings map[string]ClientLimit
	repo           *datebase.Repo
}

type ClientLimit struct {
	ID             string
	Capacity       int
	RefillInterval time.Duration
}

func NewClientManager(ctx context.Context, repo *datebase.Repo) (*ClientLimitManager, error) {
	all, err := repo.GetAll(ctx)
	if err != nil {
		slog.Error("failed to get all client limits", "error", err)
		return nil, err
	}
	clients := make(map[string]ClientLimit)
	for _, c := range all {
		clients[c.ID] = ClientLimit{
			ID:             c.ID,
			Capacity:       c.Capacity,
			RefillInterval: time.Duration(c.RefillIntervalMs) * time.Millisecond,
		}
	}

	return &ClientLimitManager{clientSettings: clients, repo: repo}, nil
}

func (m *ClientLimitManager) LookUp(ctx context.Context, id string) (ClientLimit, error) {
	limit, err := m.repo.Get(ctx, id)
	if err != nil {
		slog.Error("failed to get client limit", "id", id, "error", err)
		return ClientLimit{}, err
	}

	if limit.ID == "" {
		return ClientLimit{}, nil
	}
	return ClientLimit{
		ID:             limit.ID,
		Capacity:       limit.Capacity,
		RefillInterval: time.Duration(limit.RefillIntervalMs) * time.Millisecond,
	}, nil
}

func (m *ClientLimitManager) Get(id string) (ClientLimit, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c, ok := m.clientSettings[id]
	return c, ok
}

func (m *ClientLimitManager) Add(ctx context.Context, c ClientLimit) error {
	dbClient := datebase.ClientLimitDb{
		ID:               c.ID,
		Capacity:         c.Capacity,
		RefillIntervalMs: int64(c.RefillInterval / time.Millisecond),
	}

	err := m.repo.Add(ctx, dbClient)
	if err != nil {
		slog.Error("failed to add client limit", "id", c.ID, "error", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.clientSettings[c.ID] = c
	return nil
}

func (m *ClientLimitManager) Update(ctx context.Context, c ClientLimit) error {
	dbClient := datebase.ClientLimitDb{
		ID:               c.ID,
		Capacity:         c.Capacity,
		RefillIntervalMs: int64(c.RefillInterval / time.Millisecond),
	}

	err := m.repo.Update(ctx, dbClient)
	if err != nil {
		slog.Error("failed to update client limit", "id", c.ID, "error", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	m.clientSettings[c.ID] = c
	return nil
}

func (m *ClientLimitManager) Remove(ctx context.Context, id string) error {
	err := m.repo.Remove(ctx, id)
	if err != nil {
		slog.Error("failed to remove client limit", "id", id, "error", err)
		return err
	}

	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.clientSettings, id)
	return nil
}
