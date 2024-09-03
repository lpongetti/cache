package cache

import (
	"context"
	"time"
)

type ICache interface {
	Close() error
	Get(ctx context.Context, key string) (string, error)
	SetIfNotExists(ctx context.Context, key string, value interface{}, ttl *time.Duration) (bool, error)
	Set(ctx context.Context, key string, value interface{}, ttl *time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Increase(ctx context.Context, key string) (int64, error)
	IncreaseBy(ctx context.Context, key string, value int64) (int64, error)
	Decrease(ctx context.Context, key string) (int64, error)
	DecreaseAndDelete(ctx context.Context, key string) (int64, error)
}

type Configuration struct {
	Url string
}
