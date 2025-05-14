package ratelimit

import (
	"context"
	"time"
)

type Limiter interface {
	Allow(ctx context.Context, id string) bool
	UpdateRules(ctx context.Context, id string, newLimit int, newRefillInterval time.Duration)
	SetDefault(ctx context.Context, id string)
}
