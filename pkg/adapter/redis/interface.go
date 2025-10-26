package redis

import (
	"context"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

type Client interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	MGet(ctx context.Context, keys ...string) ([]any, error)
	MSet(ctx context.Context, values ...any) error
	Scan(ctx context.Context, cursor uint64, match string, count int64) *goredis.ScanCmd
	Keys(ctx context.Context, pattern string) *goredis.StringSliceCmd
	Exists(ctx context.Context, keys ...string) *goredis.IntCmd
	Close() error
}

type Instrument interface {
	Instrument() error
}
