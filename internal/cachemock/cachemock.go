package cachemock

import (
	"context"
	"time"

	"github.com/germandv/domainator/internal/cache"
)

type CacheMockClient struct {
	counts map[string]int
	hmap   map[string]string
}

func New() *CacheMockClient {
	return &CacheMockClient{
		counts: make(map[string]int),
	}
}

func (c *CacheMockClient) Ping() error {
	return nil
}

func (c *CacheMockClient) Close() error {
	return nil
}

func (c *CacheMockClient) Increment(key string) (int64, error) {
	c.counts[key]++
	return int64(c.counts[key]), nil
}

func (c *CacheMockClient) Expire(_ string, _ time.Duration) error {
	return nil
}

func (c *CacheMockClient) Get(_ context.Context, key string) (string, error) {
	val, ok := c.hmap[key]
	if !ok {
		return "", cache.ErrNoKey
	}
	return val, nil
}

func (c *CacheMockClient) Set(_ context.Context, key string, value string, _ time.Duration) error {
	c.hmap[key] = value
	return nil
}
