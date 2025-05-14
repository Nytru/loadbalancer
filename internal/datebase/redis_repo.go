package datebase

import (
	"cloud-test/internal/configuration"
	"context"

	"github.com/redis/go-redis/v9"
)

type RedisRepo struct {
	c *redis.Client
}

func NewRedisRepo(cfg configuration.RedisCfg) *RedisRepo {
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.Db,
	})
	return &RedisRepo{
		c: rdb,
	}
}

func (r *RedisRepo) Inc(ctx context.Context, key string, value int) error {
	return r.c.IncrBy(ctx, key, int64(value)).Err()
}

func (r *RedisRepo) Dec(ctx context.Context, key string, value int) error {
	return r.c.DecrBy(ctx, key, int64(value)).Err()
}

func (r *RedisRepo) Set(ctx context.Context, key string, value int) error {
	return r.c.Set(ctx, key, value, 0).Err()
}

func (r *RedisRepo) SetNx(ctx context.Context, key string, value int) error {
	return r.c.SetNX(ctx, key, value, 0).Err()
}

func (r *RedisRepo) Delete(ctx context.Context, key string) error {
	return r.c.Del(ctx, key).Err()
}

func (r *RedisRepo) InitAll(ctx context.Context) (map[string]int, error) {
	keys, err := r.c.Keys(ctx, "*").Result()
	if err != nil {
		return nil, err
	}
	values := make(map[string]int)
	for _, key := range keys {
		val, err := r.c.Get(ctx, key).Int()
		if err != nil {
			return nil, err
		}
		values[key] = val
	}

	return values, nil
}
