package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"api-gateway-backend/internal/config"

	"github.com/redis/go-redis/v9"
)

// Client wraps redis.Client
type Client struct {
	*redis.Client
}

// New creates a new Redis client
func New(cfg config.RedisConfig) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		DB:       cfg.DB,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Client{Client: rdb}, nil
}

// SetJSON sets a JSON value in Redis with TTL
func (c *Client) SetJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	return c.Set(ctx, key, data, ttl).Err()
}

// GetJSON gets a JSON value from Redis
func (c *Client) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.Get(ctx, key).Result()
	if err != nil {
		return err
	}

	return json.Unmarshal([]byte(data), dest)
}

// InvalidatePattern deletes all keys matching a pattern
func (c *Client) InvalidatePattern(ctx context.Context, pattern string) error {
	keys, err := c.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}

	if len(keys) > 0 {
		return c.Del(ctx, keys...).Err()
	}

	return nil
}

// Exists checks if a key exists
func (c *Client) Exists(ctx context.Context, key string) (bool, error) {
	result, err := c.Client.Exists(ctx, key).Result()
	return result > 0, err
}