package redis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/extra/redisotel/v9"
	goredis "github.com/redis/go-redis/v9"
)

type client struct {
	rdb *goredis.Client
}

func NewClient(ctx context.Context, addr string, opts ...Option) (Client, error) {
	options := &goredis.Options{
		Addr: addr,
	}

	for _, opt := range opts {
		opt(options)
	}

	rdb := goredis.NewClient(options)
	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return &client{rdb: rdb}, nil
}

func (c *client) Keys(ctx context.Context, pattern string) *goredis.StringSliceCmd {
	return c.rdb.Keys(ctx, pattern)
}

func (c *client) Scan(ctx context.Context, cursor uint64, match string, count int64) *goredis.ScanCmd {
	return c.rdb.Scan(ctx, cursor, match, count)
}

func (c *client) Exists(ctx context.Context, key ...string) *goredis.IntCmd {
	return c.rdb.Exists(ctx, key...)
}

func (c *client) Get(ctx context.Context, key string) ([]byte, error) {
	result := c.rdb.Get(ctx, key)
	if err := result.Err(); err != nil {
		return nil, err
	}

	return result.Bytes()
}

func (c *client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

func (c *client) MGet(ctx context.Context, keys ...string) ([]any, error) {
	result := c.rdb.MGet(ctx, keys...)
	if err := result.Err(); err != nil {
		return nil, err
	}

	return result.Result()
}

func (c *client) MSet(ctx context.Context, pairs ...any) error {
	return c.rdb.MSet(ctx, pairs...).Err()
}

func (c *client) Close() error {
	return c.rdb.Close()
}

func (c *client) Instrument() error {
	var joinedErr error

	if err := redisotel.InstrumentMetrics(c.rdb); err != nil {
		joinedErr = errors.Join(joinedErr, err)
	}

	if err := redisotel.InstrumentTracing(c.rdb); err != nil {
		joinedErr = errors.Join(joinedErr, err)
	}

	return joinedErr
}
