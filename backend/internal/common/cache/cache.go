package cache

import (
	"context"
	"time"
)

type Cache interface {
	Get(ctx context.Context, key string) (string, bool, error)
	Set(ctx context.Context, key string, value string, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
}

type NoopCache struct{}

func (NoopCache) Get(context.Context, string) (string, bool, error) {
	return "", false, nil
}

func (NoopCache) Set(context.Context, string, string, time.Duration) error {
	return nil
}

func (NoopCache) Delete(context.Context, string) error {
	return nil
}
