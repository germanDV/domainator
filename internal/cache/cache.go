package cache

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

var ErrNoKey = errors.New("key not found")

type Client interface {
	Ping() error
	Close() error
	Increment(key string) (int64, error)
	Expire(key string, duration time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
}

type RedisClient struct {
	client *redis.Client
}

func (rc *RedisClient) Ping() error {
	resp, err := rc.client.Ping(context.Background()).Result()
	if err != nil || resp != "PONG" {
		return err
	}
	return nil
}

func (rc *RedisClient) Close() error {
	return rc.client.Close()
}

func (rc *RedisClient) Increment(key string) (int64, error) {
	return rc.client.Incr(context.Background(), key).Result()
}

func (rc *RedisClient) Expire(key string, duration time.Duration) error {
	return rc.client.Expire(context.Background(), key, duration).Err()
}

func (rc *RedisClient) Get(ctx context.Context, key string) (string, error) {
	val, err := rc.client.Get(ctx, key).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return "", ErrNoKey
		}
		return "", err
	}
	return val, nil
}

func (rc *RedisClient) Set(ctx context.Context, key string, value string, ttl time.Duration) error {
	return rc.client.Set(ctx, key, value, ttl).Err()
}

func New(host string, port int, password string) Client {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       0,
		}),
	}
}
