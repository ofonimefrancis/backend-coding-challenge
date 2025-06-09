package cache

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisConfig struct {
	Host         string        `env:"REDIS_HOST" default:"localhost"`
	Port         int           `env:"REDIS_PORT" default:"6379"`
	Password     string        `env:"REDIS_PASSWORD" default:""`
	DB           int           `env:"REDIS_DB" default:"0"`
	MaxRetries   int           `env:"REDIS_MAX_RETRIES" default:"3"`
	PoolSize     int           `env:"REDIS_POOL_SIZE" default:"10"`
	MinIdleConns int           `env:"REDIS_MIN_IDLE_CONNS" default:"5"`
	DialTimeout  time.Duration `env:"REDIS_DIAL_TIMEOUT" default:"5s"`
	ReadTimeout  time.Duration `env:"REDIS_READ_TIMEOUT" default:"3s"`
	WriteTimeout time.Duration `env:"REDIS_WRITE_TIMEOUT" default:"3s"`
	PoolTimeout  time.Duration `env:"REDIS_POOL_TIMEOUT" default:"4s"`
	IdleTimeout  time.Duration `env:"REDIS_IDLE_TIMEOUT" default:"5m"`
}

type Cache interface {
	Get(ctx context.Context, key string, dest interface{}) error
	Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
	Delete(ctx context.Context, keys ...string) error
	DeletePattern(ctx context.Context, pattern string) error
	Exists(ctx context.Context, key string) (bool, error)
	TTL(ctx context.Context, key string) (time.Duration, error)
	Close() error

	// Batch operations
	MGet(ctx context.Context, keys []string, dest interface{}) error
	MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error

	// Health check
	Ping(ctx context.Context) error
}

type redisCache struct {
	client *redis.Client
	prefix string
}

func NewRedisCache(config RedisConfig, prefix string) (Cache, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password:     config.Password,
		DB:           config.DB,
		MaxRetries:   config.MaxRetries,
		PoolSize:     config.PoolSize,
		MinIdleConns: config.MinIdleConns,
		DialTimeout:  config.DialTimeout,
		ReadTimeout:  config.ReadTimeout,
		WriteTimeout: config.WriteTimeout,
		PoolTimeout:  config.PoolTimeout,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &redisCache{
		client: rdb,
		prefix: prefix,
	}, nil
}

func (r *redisCache) getKey(key string) string {
	if r.prefix == "" {
		return key
	}
	return fmt.Sprintf("%s:%s", r.prefix, key)
}

func (r *redisCache) Get(ctx context.Context, key string, dest interface{}) error {
	val, err := r.client.Get(ctx, r.getKey(key)).Result()
	if err != nil {
		if err == redis.Nil {
			return ErrCacheMiss
		}
		return fmt.Errorf("redis get error: %w", err)
	}

	if err := json.Unmarshal([]byte(val), dest); err != nil {
		return fmt.Errorf("json unmarshal error: %w", err)
	}

	return nil
}

func (r *redisCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("json marshal error: %w", err)
	}

	if err := r.client.Set(ctx, r.getKey(key), data, ttl).Err(); err != nil {
		return fmt.Errorf("redis set error: %w", err)
	}

	return nil
}

func (r *redisCache) Delete(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.getKey(key)
	}

	if err := r.client.Del(ctx, prefixedKeys...).Err(); err != nil {
		return fmt.Errorf("redis delete error: %w", err)
	}

	return nil
}

func (r *redisCache) DeletePattern(ctx context.Context, pattern string) error {
	fullPattern := r.getKey(pattern)

	iter := r.client.Scan(ctx, 0, fullPattern, 0).Iterator()
	var keys []string

	for iter.Next(ctx) {
		keys = append(keys, iter.Val())

		// Delete in batches of 100
		if len(keys) >= 100 {
			if err := r.client.Del(ctx, keys...).Err(); err != nil {
				return fmt.Errorf("redis batch delete error: %w", err)
			}
			keys = keys[:0]
		}
	}

	// Delete remaining keys
	if len(keys) > 0 {
		if err := r.client.Del(ctx, keys...).Err(); err != nil {
			return fmt.Errorf("redis final delete error: %w", err)
		}
	}

	return iter.Err()
}

func (r *redisCache) Exists(ctx context.Context, key string) (bool, error) {
	count, err := r.client.Exists(ctx, r.getKey(key)).Result()
	if err != nil {
		return false, fmt.Errorf("redis exists error: %w", err)
	}
	return count > 0, nil
}

func (r *redisCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	ttl, err := r.client.TTL(ctx, r.getKey(key)).Result()
	if err != nil {
		return 0, fmt.Errorf("redis ttl error: %w", err)
	}
	return ttl, nil
}

func (r *redisCache) MGet(ctx context.Context, keys []string, dest interface{}) error {
	if len(keys) == 0 {
		return nil
	}

	prefixedKeys := make([]string, len(keys))
	for i, key := range keys {
		prefixedKeys[i] = r.getKey(key)
	}

	vals, err := r.client.MGet(ctx, prefixedKeys...).Result()
	if err != nil {
		return fmt.Errorf("redis mget error: %w", err)
	}

	// Convert results to proper format
	results := make(map[string]interface{})
	for i, val := range vals {
		if val != nil {
			var data interface{}
			if err := json.Unmarshal([]byte(val.(string)), &data); err == nil {
				results[keys[i]] = data
			}
		}
	}

	// Marshal and unmarshal to convert to dest type
	resultData, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("json marshal results error: %w", err)
	}

	return json.Unmarshal(resultData, dest)
}

func (r *redisCache) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	if len(items) == 0 {
		return nil
	}

	pipe := r.client.Pipeline()

	for key, value := range items {
		data, err := json.Marshal(value)
		if err != nil {
			return fmt.Errorf("json marshal error for key %s: %w", key, err)
		}
		pipe.Set(ctx, r.getKey(key), data, ttl)
	}

	_, err := pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("redis pipeline exec error: %w", err)
	}

	return nil
}

func (r *redisCache) Ping(ctx context.Context) error {
	return r.client.Ping(ctx).Err()
}

func (r *redisCache) Close() error {
	return r.client.Close()
}

// Custom errors
var (
	ErrCacheMiss = fmt.Errorf("cache miss")
)

// NoOpCache is a no-operation cache implementation for non-production environments
// It always returns cache misses and does not store any data.
type NoOpCache struct{}

func NewNoOpCache() Cache {
	return &NoOpCache{}
}

func (n *NoOpCache) Get(ctx context.Context, key string, dest interface{}) error {
	return ErrCacheMiss
}

func (n *NoOpCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return nil
}

func (n *NoOpCache) Delete(ctx context.Context, keys ...string) error {
	return nil
}

func (n *NoOpCache) DeletePattern(ctx context.Context, pattern string) error {
	return nil
}

func (n *NoOpCache) Exists(ctx context.Context, key string) (bool, error) {
	return false, nil
}

func (n *NoOpCache) TTL(ctx context.Context, key string) (time.Duration, error) {
	return 0, nil
}

func (n *NoOpCache) Close() error {
	return nil
}

func (n *NoOpCache) MGet(ctx context.Context, keys []string, dest interface{}) error {
	return nil
}

func (n *NoOpCache) MSet(ctx context.Context, items map[string]interface{}, ttl time.Duration) error {
	return nil
}

func (n *NoOpCache) Ping(ctx context.Context) error {
	return nil
}
