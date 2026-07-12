package cache

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrCacheMiss = errors.New("cache miss")

type Cache interface {
	Get(ctx context.Context, key string, dest any) error
	Set(ctx context.Context, key string, value any, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeleteByPattern(ctx context.Context, pattern string) error
	Ping(ctx context.Context) error
	Close() error
}

const maxBackgroundTasks = 32

var backgroundTasks = make(chan struct{}, maxBackgroundTasks)

// InBackground запускает необязательную работу с кэшем, не блокируя HTTP-ответ.
// Если очередь занята, задача пропускается: кэш будет заполнен следующим запросом.
func InBackground(fn func()) {
	select {
	case backgroundTasks <- struct{}{}:
		go func() {
			defer func() { <-backgroundTasks }()
			fn()
		}()
	default:
	}
}

type Config struct {
	Addr     string
	Password string
	DB       int
	Disabled bool
}

type redisCache struct {
	client   *redis.Client
	disabled bool
}

func NewRedisCache(cfg Config) (Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password,
		DB:       cfg.DB,

		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,

		PoolSize:     10,
		MinIdleConns: 3,
	})

	if !cfg.Disabled {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := client.Ping(ctx).Err(); err != nil {
			return nil, fmt.Errorf("ping redis: %w", err)
		}
	}

	return &redisCache{client: client, disabled: cfg.Disabled}, nil
}

func (c *redisCache) Get(ctx context.Context, key string, dest any) error {
	if c.disabled {
		return ErrCacheMiss
	}

	data, err := c.client.Get(ctx, key).Bytes()

	if err != nil {
		if errors.Is(err, redis.Nil) {
			return ErrCacheMiss
		}
		return fmt.Errorf("cache get %q: %w", key, err)
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return fmt.Errorf("cache unmarshal %q: %w", key, err)
	}

	return nil
}

func (c *redisCache) Set(ctx context.Context, key string, value any, ttl time.Duration) error {
	if c.disabled {
		return nil
	}

	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("cache marshal %q: %w", key, err)
	}

	if err := c.client.Set(ctx, key, data, ttl).Err(); err != nil {
		return fmt.Errorf("cache set %q: %w", key, err)
	}

	return nil
}

func (c *redisCache) Delete(ctx context.Context, keys ...string) error {
	if c.disabled {
		return nil
	}

	if err := c.client.Del(ctx, keys...).Err(); err != nil {
		return fmt.Errorf("cache delete: %w", err)
	}
	return nil
}

func (c *redisCache) DeleteByPattern(ctx context.Context, pattern string) error {
	if c.disabled {
		return nil
	}

	var cursor uint64
	for {
		keys, nextCursor, err := c.client.Scan(ctx, cursor, pattern, 100).Result()
		if err != nil {
			return fmt.Errorf("scan pattern %q: %w", pattern, err)
		}

		if len(keys) > 0 {
			if err := c.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("delete keys: %w", err)
			}
		}

		cursor = nextCursor
		if cursor == 0 {
			break
		}
	}

	return nil
}

func (c *redisCache) Ping(ctx context.Context) error {
	if c.disabled {
		return nil
	}

	if err := c.client.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis ping: %w", err)
	}
	return nil
}

func (c *redisCache) Close() error {
	return c.client.Close()
}
