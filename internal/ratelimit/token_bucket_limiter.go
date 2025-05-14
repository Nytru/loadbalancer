package ratelimit

import (
	"cloud-test/internal/clients"
	"cloud-test/internal/configuration"
	"cloud-test/internal/datebase"
	"context"
	"sync"
	"time"
)

var _ Limiter = (*TokenBucketLimiter)(nil)

type TokenBucketLimiter struct {
	enabled         bool
	redisRepo       *datebase.RedisRepo
	clientMgr       *clients.ClientLimitManager
	buckets         map[string]*Bucket
	defaultCfg      limiterConfig
	mu              sync.RWMutex
	appCtx          context.Context
}

type limiterConfig struct {
	maxCapacity    int
	refillInterval time.Duration
}

type Bucket struct {
	ID               string
	currentLimitLeft int
	cfg              limiterConfig
	mu               sync.RWMutex
	lastUse          time.Time
}

func NewLimiterManager(
	appCtx context.Context,
	redisRepo *datebase.RedisRepo,
	clientMgr *clients.ClientLimitManager,
	defaultCfg configuration.RateCfg) (*TokenBucketLimiter, error) {

	restoredLimits, err := redisRepo.InitAll(appCtx)
	if err != nil {
		return nil, err
	}

	lm := &TokenBucketLimiter{
		redisRepo:       redisRepo,
		clientMgr:       clientMgr,
		buckets:         make(map[string]*Bucket),
		defaultCfg: limiterConfig{
			maxCapacity:    defaultCfg.Capacity,
			refillInterval: defaultCfg.RefillInterval,
		},
		enabled: defaultCfg.Enabled,
		appCtx:  appCtx,
	}
	for id, limitLeft := range restoredLimits {
		clientConfig, exists := clientMgr.Get(id)
		if !exists {
			clientConfig = clients.ClientLimit{
				ID:             id,
				Capacity:       defaultCfg.Capacity,
				RefillInterval: defaultCfg.RefillInterval,
			}
		}

		if limitLeft > clientConfig.Capacity {
			limitLeft = clientConfig.Capacity
		}

		bucket := &Bucket{
			ID: id,
			cfg: limiterConfig{
				maxCapacity:    clientConfig.Capacity,
				refillInterval: clientConfig.RefillInterval,
			},
			currentLimitLeft: limitLeft,
			lastUse:          time.Now(),
		}
		_ = lm.redisRepo.SetNx(appCtx, id, clientConfig.Capacity)
		go lm.refillBucket(bucket)

		lm.buckets[id] = bucket
	}

	return lm, nil
}

func (lm *TokenBucketLimiter) Allow(ctx context.Context, id string) bool {
	if !lm.enabled {
		return true
	}

	lm.mu.RLock()
	bucket, exists := lm.buckets[id]
	lm.mu.RUnlock()

	if !exists {
		clientConfig, exists := lm.clientMgr.Get(id)
		if !exists {
			clientConfig = clients.ClientLimit{
				ID:             id,
				Capacity:       lm.defaultCfg.maxCapacity,
				RefillInterval: lm.defaultCfg.refillInterval,
			}
		}

		bucket = &Bucket{
			ID: id,
			cfg: limiterConfig{
				maxCapacity:    clientConfig.Capacity,
				refillInterval: clientConfig.RefillInterval,
			},
			currentLimitLeft: clientConfig.Capacity,
			lastUse:          time.Now(),
		}
		lm.mu.Lock()
		lm.buckets[id] = bucket
		lm.mu.Unlock()

		_ = lm.redisRepo.SetNx(ctx, id, clientConfig.Capacity)
		go lm.refillBucket(bucket)
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.lastUse = time.Now()
	if bucket.currentLimitLeft > 0 {
		bucket.currentLimitLeft--
		lm.redisRepo.Dec(ctx, id, 1)
		return true
	}

	return false
}

func (lm *TokenBucketLimiter) UpdateRules(ctx context.Context, id string, newCapacity int, newRefillInterval time.Duration) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	bucket, exists := lm.buckets[id]
	if !exists {
		bucket = &Bucket{
			ID:               id,
			currentLimitLeft: newCapacity,
			cfg: limiterConfig{
				maxCapacity:    newCapacity,
				refillInterval: newRefillInterval,
			},
			lastUse: time.Now(),
		}
		_ = lm.redisRepo.SetNx(ctx, id, newCapacity)
		go lm.refillBucket(bucket)
		lm.buckets[id] = bucket
		return
	}

	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.cfg = limiterConfig{
		maxCapacity:    newCapacity,
		refillInterval: newRefillInterval,
	}

	if bucket.currentLimitLeft > newCapacity {
		lm.redisRepo.Set(ctx, id, newCapacity)
		bucket.currentLimitLeft = newCapacity
	}
}

func (lm *TokenBucketLimiter) SetDefault(ctx context.Context, id string) {
	lm.mu.Lock()
	defer lm.mu.Unlock()

	bucket, exists := lm.buckets[id]
	if !exists {
		return
	}
	bucket.mu.Lock()
	defer bucket.mu.Unlock()

	bucket.cfg = limiterConfig{
		maxCapacity:    lm.defaultCfg.maxCapacity,
		refillInterval: lm.defaultCfg.refillInterval,
	}

	if bucket.currentLimitLeft > lm.defaultCfg.maxCapacity {
		lm.redisRepo.Set(ctx, id, lm.defaultCfg.maxCapacity)
		bucket.currentLimitLeft = lm.defaultCfg.maxCapacity
	}
}

func (lm *TokenBucketLimiter) refillBucket(bucket *Bucket) {
	ticker := time.NewTicker(bucket.cfg.refillInterval)
	defer ticker.Stop()

	for {
		select {
		case <-lm.appCtx.Done():
			return
		case <-ticker.C:
			ticker.Reset(bucket.cfg.refillInterval)
			bucket.mu.Lock()
			if bucket.cfg.maxCapacity > bucket.currentLimitLeft {
				bucket.currentLimitLeft++
				lm.redisRepo.Inc(lm.appCtx, bucket.ID, 1)
			}
			bucket.mu.Unlock()
		}
	}
}
