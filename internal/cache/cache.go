package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	Ping() error
	Close() error
	Increment(key string) (int64, error)
	Expire(key string, duration time.Duration) error
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

func New(host string, port int, password string) Client {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       0,
		}),
	}
}
