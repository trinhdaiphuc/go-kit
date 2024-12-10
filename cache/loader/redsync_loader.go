package cacheloader

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/cache"
	redislock "github.com/trinhdaiphuc/go-kit/cache/redis/lock"
	"github.com/trinhdaiphuc/go-kit/log"
)

type RedSyncLoader[K comparable, V any] struct {
	redLock redislock.RedLock
	loader  cache.Loader[K, V]
	expiry  time.Duration
	loadKey LoadKeyFunc
}

type LoadKeyFunc func(string) string

func NewRedSyncLoader[K comparable, V any](redLock redislock.RedLock, loader cache.Loader[K, V], loadKey LoadKeyFunc, expiry time.Duration) *RedSyncLoader[K, V] {
	return &RedSyncLoader[K, V]{
		redLock: redLock,
		loader:  loader,
		expiry:  expiry,
		loadKey: loadKey,
	}
}

func (r *RedSyncLoader[K, V]) Load(ctx context.Context, c cache.Store[K, V], key K) (value V, err error) {
	mutex := r.redLock.GetLock(r.loadKey(defaultKeyEncoder(key)), r.expiry)
	err = mutex.TryLockContext(ctx)
	if err != nil {
		return value, fmt.Errorf("acquire lock failed: %w", err)
	}

	defer func() {
		ok, errUnlock := mutex.Unlock()
		if !ok || errUnlock != nil {
			log.For(ctx).Error("Unlock failed", zap.Error(err))
		}
	}()

	value, err = r.loader.Load(ctx, c, key)

	return
}

func (r *RedSyncLoader[K, V]) LoadAll(ctx context.Context, c cache.Store[K, V], key K) (map[K]V, error) {
	mutex := r.redLock.GetLock(defaultKeyEncoder(key), r.expiry)
	err := mutex.TryLockContext(ctx)
	if err != nil {
		return nil, fmt.Errorf("acquire lock failed: %w", err)
	}

	defer func() {
		ok, errUnlock := mutex.Unlock()
		if !ok || errUnlock != nil {
			log.For(ctx).Error("Unlock failed", zap.Error(err))
		}
	}()

	return r.loader.LoadAll(ctx, c, key)
}

func defaultKeyEncoder(key any) string {
	return fmt.Sprint(key)
}
