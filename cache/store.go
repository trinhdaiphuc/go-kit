package cache

import (
	"context"
	"errors"
	"time"
)

type KeyVal[K comparable, V any] struct {
	Key   K
	Value V
}

//go:generate mockgen -destination=./mocks/$GOFILE -source=$GOFILE -package=cachemock
type Store[K comparable, V any] interface {
	Get(ctx context.Context, key K) (V, error)
	Set(ctx context.Context, key K, value V) error
	SetNX(ctx context.Context, key K, value V) (bool, error)
	BulkGet(ctx context.Context, keys []K) (map[K]V, error)
	Delete(ctx context.Context, keys ...K) error
	Incr(ctx context.Context, key K, value int64) (int64, error)
	Expire(ctx context.Context, key K, expireTime time.Duration) error
	TTL(ctx context.Context, key K) (time.Duration, error)
	HSet(ctx context.Context, key K, keyVals ...KeyVal[K, V]) error
	HGet(ctx context.Context, key, field K) (V, error)
	HGetAll(ctx context.Context, key K) (map[K]V, error)
	HDel(ctx context.Context, key K, fields ...K) error
	Ping(ctx context.Context) error
	Close()
}

// Loader is an interface that handles missing data loading.
type Loader[K comparable, V any] interface {
	// Load should execute a custom item retrieval logic and
	// return the item that is associated with the key.
	// It should return nil if the item is not found/valid.
	// The method is allowed to fetch data from the cache instance
	// or update it for future use.
	Load(ctx context.Context, c Store[K, V], key K) (value V, err error)

	// LoadAll should execute a custom item retrieval logic and
	// return the items that are associated with the keys.
	// It should return nil if the item is not found/valid.
	// The method is allowed to fetch data from the cache instance
	// or update it for future use.
	LoadAll(ctx context.Context, c Store[K, V], key K) (map[K]V, error)

	// BulkLoad computes or retrieves the values corresponding to keys.
	// This method is called by Cache.BulkGet.
	//
	// If the returned map doesn't contain all requested keys, then the entries it does
	// contain will be cached, and Cache.BulkGet will return the partial results. If the returned map
	// contains extra keys not present in keys then all returned entries will be cached, but
	// only the entries for keys, will be returned from Cache.BulkGet.
	//
	// WARNING: loading must not attempt to update any mappings of this cache directly.
	BulkLoad(ctx context.Context, c Store[K, V], keys []K) (map[K]V, error)
}

// LoaderFunc type is an adapter that allows the use of ordinary
// functions as data loaders.
type LoaderFunc[K comparable, V any] func(ctx context.Context, c Store[K, V], key K) (value V, err error)

// LoaderFuncAll type is an adapter that allows the use of ordinary
// functions as data loaders.
type LoaderFuncAll[K comparable, V any] func(ctx context.Context, c Store[K, V], key K) (map[K]V, error)

// BulkLoaderFunc type is an adapter that allows the use of ordinary
// functions as data loaders for multiple keys.
type BulkLoaderFunc[K comparable, V any] func(ctx context.Context, c Store[K, V], keys []K) (map[K]V, error)

// Load executes a custom item retrieval logic and returns the item that
// is associated with the key.
// It returns nil if the item is not found/valid.
func (l LoaderFunc[K, V]) Load(ctx context.Context, c Store[K, V], key K) (value V, err error) {
	return l(ctx, c, key)
}

// LoadAll executes a custom item retrieval logic and returns the items that
// are associated with the keys.
// It returns nil if the item is not found/valid.
func (l LoaderFuncAll[K, V]) LoadAll(ctx context.Context, c Store[K, V], key K) (map[K]V, error) {
	return l(ctx, c, key)
}

// BulkLoad computes or retrieves the values corresponding to keys.
func (l BulkLoaderFunc[K, V]) BulkLoad(ctx context.Context, c Store[K, V], keys []K) (map[K]V, error) {
	if len(keys) == 0 {
		return nil, nil
	}

	return l(ctx, c, keys)
}

var (
	ErrorKeyNotFound    = errors.New("key not found")
	ErrorFailedSetCache = errors.New("failed to set cache")
)
