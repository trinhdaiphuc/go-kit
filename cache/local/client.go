package cachelocal

import (
	"context"
	"time"

	"github.com/jellydator/ttlcache/v3"

	"github.com/trinhdaiphuc/go-kit/cache"
)

type client[K comparable, V any] struct {
	cli  *ttlcache.Cache[K, V]
	opts *Options[K, V]
}

func NewClient[K comparable, V any](opts ...Option[K, V]) cache.Store[K, V] {
	option := defaultOption[K, V]()

	for _, opt := range opts {
		opt(option)
	}

	cli := &client[K, V]{
		cli: ttlcache.New[K, V](
			ttlcache.WithDisableTouchOnHit[K, V](),
			ttlcache.WithTTL[K, V](option.TTL),
		),
		opts: option,
	}

	go cli.cleanUpExpired()

	return cli
}

func (c *client[K, V]) Get(ctx context.Context, key K) (v V, err error) {
	item := c.cli.Get(key, ttlcache.WithLoader[K, V](WrapLoadFunc[K, V](c.opts, ctx, c, key)))
	if item == nil {
		return v, cache.ErrorKeyNotFound
	}

	return item.Value(), nil
}

func (c *client[K, V]) BulkGet(ctx context.Context, keys []K) (map[K]V, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	result := make(map[K]V, len(keys))
	for _, key := range keys {
		v, err := c.Get(ctx, key)
		if err != nil {
			continue
		}
		result[key] = v
	}

	return result, nil
}

func (c *client[K, V]) Set(ctx context.Context, key K, value V) error {
	item := c.cli.Set(key, value, c.opts.TTL)
	if item == nil {
		return cache.ErrorFailedSetCache
	}

	return nil
}

func (c *client[K, V]) SetNX(ctx context.Context, key K, value V) (bool, error) {
	// TTLCache doesn't have native SetNX, so we check existence first
	if c.cli.Has(key) {
		return false, nil
	}
	item := c.cli.Set(key, value, c.opts.TTL)
	if item == nil {
		return false, cache.ErrorFailedSetCache
	}
	return true, nil
}

func (c *client[K, V]) Delete(ctx context.Context, keys ...K) error {
	for _, key := range keys {
		c.cli.Delete(key)
	}
	return nil
}

func (c *client[K, V]) Incr(ctx context.Context, key K, value int64) (int64, error) {
	// TTLCache doesn't support atomic increment
	return 0, nil
}

func (c *client[K, V]) Expire(ctx context.Context, key K, expireTime time.Duration) error {
	item := c.cli.Get(key)
	if item == nil {
		return cache.ErrorKeyNotFound
	}
	c.cli.Set(key, item.Value(), expireTime)
	return nil
}

func (c *client[K, V]) TTL(ctx context.Context, key K) (time.Duration, error) {
	item := c.cli.Get(key)
	if item == nil {
		return 0, cache.ErrorKeyNotFound
	}
	return time.Until(item.ExpiresAt()), nil
}

func (c *client[K, V]) HSet(ctx context.Context, key K, keyVals ...cache.KeyVal[K, V]) error {
	return nil
}

func (c *client[K, V]) HGet(ctx context.Context, key, field K) (v V, err error) {
	return
}

func (c *client[K, V]) HGetAll(ctx context.Context, key K) (map[K]V, error) {
	return nil, nil
}

func (c *client[K, V]) HDel(ctx context.Context, key K, fields ...K) error {
	return nil
}

func (c *client[K, V]) Ping(ctx context.Context) error {
	return nil
}

func (c *client[K, V]) Close() {
	c.cli.Stop()
}

func (c *client[K, V]) cleanUpExpired() {
	for {
		time.Sleep(c.opts.CleanUpInterval)
		c.cli.DeleteExpired()
	}
}

func WrapLoadFunc[K comparable, V any](opts *Options[K, V], ctx context.Context, store cache.Store[K, V], key K) ttlcache.LoaderFunc[K, V] {
	return func(ttlCache *ttlcache.Cache[K, V], key K) *ttlcache.Item[K, V] {
		if opts.Loader == nil {
			return nil
		}
		value, err := opts.Loader.Load(ctx, store, key)
		if err != nil {
			return nil
		}

		return ttlCache.Set(key, value, opts.TTL)
	}
}
