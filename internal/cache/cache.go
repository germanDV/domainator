package cache

import (
	"context"
	"fmt"

	"github.com/redis/go-redis/v9"
)

type Client interface {
	Ping() error
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

func New(host string, port int, password string) Client {
	return &RedisClient{
		client: redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%d", host, port),
			Password: password,
			DB:       0,
		}),
	}
}
