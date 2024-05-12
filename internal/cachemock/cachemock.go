package cachemock

import "time"

type CacheMockClient struct {
	counts map[string]int
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
