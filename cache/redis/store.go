package cacheredis

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/cache"
	"github.com/trinhdaiphuc/go-kit/log"
)

type redisCache[K comparable, V any] struct {
	client redis.UniversalClient
	opts   *Options[K, V]
}

func NewRedisCache[K comparable, V any](cli redis.UniversalClient, options ...Option[K, V]) cache.Store[K, V] {
	opts := newDefaultOption[K, V]()
	for _, o := range options {
		o(opts)
	}

	return &redisCache[K, V]{
		client: cli,
		opts:   opts,
	}
}

func (c *redisCache[K, V]) Get(ctx context.Context, key K) (value V, err error) {
	data, err := c.client.Get(ctx, c.encodeKey(key)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return c.load(ctx, key)
		}
		return value, err
	}

	err = c.unmarshal(data, &value)

	return
}

func (c *redisCache[K, V]) load(ctx context.Context, key K) (value V, err error) {
	if c.opts == nil || c.opts.Loader == nil {
		return value, cache.ErrorKeyNotFound
	}

	value, err = c.opts.Loader.Load(ctx, c, key)
	if err != nil {
		return value, err
	}

	return value, nil
}

func (c *redisCache[K, V]) loadAll(ctx context.Context, key K) (map[K]V, error) {
	if c.opts == nil || c.opts.Loader == nil {
		return nil, cache.ErrorKeyNotFound
	}

	value, err := c.opts.Loader.LoadAll(ctx, c, key)
	if err != nil {
		return value, err
	}

	return value, nil
}

func (c *redisCache[K, V]) Set(ctx context.Context, key K, value V, expireSecond time.Duration) error {
	data, err := c.marshal(value)
	if err != nil {
		return err
	}
	return c.client.Set(ctx, c.encodeKey(key), data, expireSecond).Err()
}

func (c *redisCache[K, V]) SetNX(ctx context.Context, key K, value V, expireSecond time.Duration) (bool, error) {
	data, err := c.marshal(value)
	if err != nil {
		return false, err
	}
	return c.client.SetNX(ctx, c.encodeKey(key), data, expireSecond).Result()
}

func (c *redisCache[K, V]) Delete(ctx context.Context, keys ...K) error {
	keyVals := make([]string, 0, len(keys))
	for _, key := range keys {
		keyVals = append(keyVals, c.encodeKey(key))
	}
	return c.client.Del(ctx, keyVals...).Err()
}

func (c *redisCache[K, V]) Incr(ctx context.Context, key K, value int64) (int64, error) {
	return c.client.IncrBy(ctx, c.encodeKey(key), value).Result()
}

func (c *redisCache[K, V]) Expire(ctx context.Context, key K, expireTime time.Duration) error {
	return c.client.Expire(ctx, c.encodeKey(key), expireTime).Err()
}

func (c *redisCache[K, V]) TTL(ctx context.Context, key K) (time.Duration, error) {
	return c.client.TTL(ctx, c.encodeKey(key)).Result()
}

func (c *redisCache[K, V]) HSet(ctx context.Context, key K, keyVals ...cache.KeyVal[K, V]) error {
	values := make([]interface{}, 0, len(keyVals)*2)
	for _, keyVal := range keyVals {
		data, err := c.marshal(keyVal.Value)
		if err != nil {
			return err
		}
		values = append(values, c.opts.KeyEncoder(keyVal.Key), data)
	}

	return c.client.HSet(ctx, c.encodeKey(key), values...).Err()
}

func (c *redisCache[K, V]) HGet(ctx context.Context, key, field K) (value V, err error) {
	data, err := c.client.HGet(ctx, c.encodeKey(key), c.opts.KeyEncoder(field)).Result()
	if err != nil {
		if errors.Is(err, redis.Nil) {
			return c.load(ctx, field)
		}
		return value, err
	}

	err = c.unmarshal(data, &value)
	return
}

func (c *redisCache[K, V]) HGetAll(ctx context.Context, key K) (map[K]V, error) {
	result, err := c.client.HGetAll(ctx, c.encodeKey(key)).Result()
	if err != nil {
		return nil, err
	}

	// HGETALL returns an empty map if the key does not exist
	if len(result) == 0 {
		return c.loadAll(ctx, key)
	}

	rs := make(map[K]V, len(result))
	for keyStr, value := range result {
		var data V
		err := c.unmarshal(value, &data)
		if err != nil {
			return nil, err
		}
		rs[c.decodeHashKey(keyStr)] = data
	}

	return rs, nil
}

func (c *redisCache[K, V]) HDel(ctx context.Context, key K, fields ...K) error {
	fieldVals := make([]string, 0, len(fields))
	for _, field := range fields {
		fieldVals = append(fieldVals, c.opts.KeyEncoder(field))
	}
	return c.client.HDel(ctx, c.encodeKey(key), fieldVals...).Err()
}

func (c *redisCache[K, V]) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}

func (c *redisCache[K, V]) Close() {
	err := c.client.Close()
	if err != nil {
		log.Bg().Error("Error closing redis client", zap.Error(err))
		return
	}
	log.Bg().Info("redis client closed")
}

func (c *redisCache[K, V]) marshal(value V) (string, error) {
	data, err := c.opts.MarshalValue(value)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

func (c *redisCache[K, V]) unmarshal(data string, value *V) (err error) {
	err = c.opts.UnmarshalValue([]byte(data), value)
	return
}

func (c *redisCache[K, V]) encodeKey(key K) string {
	if c.opts.Prefix == "" {
		return c.opts.KeyEncoder(key)
	}

	return c.opts.Prefix + ":" + c.opts.KeyEncoder(key)
}

func (c *redisCache[K, V]) decodeHashKey(key string) (result K) {
	k := c.opts.KeyDecoder(key)

	decodeKey, ok := k.(K)
	if !ok {
		return
	}
	return decodeKey
}
