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

func (c *client[K, V]) Set(ctx context.Context, key K, value V, expireSecond time.Duration) error {
	item := c.cli.Set(key, value, expireSecond)
	if item == nil {
		return cache.ErrorKeyNotFound
	}

	return nil
}

func (c *client[K, V]) SetNX(ctx context.Context, key K, value V, expireSecond time.Duration) (bool, error) {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) Delete(ctx context.Context, keys ...K) error {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) Incr(ctx context.Context, key K, value int64) (int64, error) {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) Expire(ctx context.Context, key K, expireTime time.Duration) error {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) TTL(ctx context.Context, key K) (time.Duration, error) {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) HSet(ctx context.Context, key K, keyVals ...cache.KeyVal[K, V]) error {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) HGet(ctx context.Context, key, field K) (V, error) {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) HGetAll(ctx context.Context, key K) (map[K]V, error) {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) HDel(ctx context.Context, key K, fields ...K) error {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) Ping(ctx context.Context) error {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) Close() {
	// TODO implement me
	panic("implement me")
}

func (c *client[K, V]) cleanUpExpired() {
	for {
		time.Sleep(c.opts.CleanUpInterval)
		c.cli.DeleteExpired()
	}
}

func WrapLoadFunc[K comparable, V any](opts *Options[K, V], ctx context.Context, store cache.Store[K, V], key K) ttlcache.LoaderFunc[K, V] {
	return func(ttlCache *ttlcache.Cache[K, V], key K) *ttlcache.Item[K, V] {
		value, err := opts.Loader.Load(ctx, store, key)
		if err != nil {
			return nil
		}

		return ttlCache.Set(key, value, opts.TTL)
	}
}
