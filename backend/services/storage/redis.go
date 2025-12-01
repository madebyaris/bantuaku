package storage

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Redis wraps a Redis client
type Redis struct {
	client *redis.Client
}

// NewRedis creates a new Redis client
func NewRedis(redisURL string) (*Redis, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}

	client := redis.NewClient(opt)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &Redis{client: client}, nil
}

// Close closes the Redis client
func (r *Redis) Close() error {
	return r.client.Close()
}

// Client returns the underlying Redis client
func (r *Redis) Client() *redis.Client {
	return r.client
}

// Get retrieves a value from cache
func (r *Redis) Get(ctx context.Context, key string) (string, error) {
	return r.client.Get(ctx, key).Result()
}

// Set stores a value in cache with TTL
func (r *Redis) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	return r.client.Set(ctx, key, value, ttl).Err()
}

// Delete removes a key from cache
func (r *Redis) Delete(ctx context.Context, keys ...string) error {
	return r.client.Del(ctx, keys...).Err()
}

// Exists checks if a key exists
func (r *Redis) Exists(ctx context.Context, key string) (bool, error) {
	n, err := r.client.Exists(ctx, key).Result()
	return n > 0, err
}
