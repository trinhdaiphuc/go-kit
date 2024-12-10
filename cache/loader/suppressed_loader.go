package cacheloader

import (
	"context"
	"fmt"

	"golang.org/x/sync/singleflight"

	"github.com/trinhdaiphuc/go-kit/cache"
)

// SuppressedLoader wraps another Loader and suppresses duplicate
// calls to its Load method.
type SuppressedLoader[K comparable, V any] struct {
	loader cache.Loader[K, V]
	group  *singleflight.Group
}

// Load executes a custom item retrieval logic and returns the item that
// is associated with the key.
// It returns nil if the item is not found/valid.
// It also ensures that only one execution of the wrapped Loader's Load
// method is in-flight for a given key at a time.
func (l *SuppressedLoader[K, V]) Load(ctx context.Context, c cache.Store[K, V], key K) (value V, err error) {
	// there should be a better/generic way to create a
	// singleflight Group's key. It's possible that a generic
	// singleflight.Group will be introduced with/in go1.19+
	strKey := defaultKeyEncoder(key)

	// the error can be discarded since the singleflight.Group
	// itself does not return any of its errors, it returns
	// the error that we return ourselves in the func below, which
	// is also nil
	res, err, _ := l.group.Do(strKey, func() (interface{}, error) {
		v, err := l.loader.Load(ctx, c, key)
		if err != nil {
			return nil, err
		}

		return v, nil
	})
	if err != nil {
		return value, err
	}

	if res == nil {
		return value, nil
	}

	var ok bool
	value, ok = res.(V)
	if !ok {
		return value, fmt.Errorf("invalid type %T, expected %T", res, value)
	}

	return value, nil
}

// LoadAll executes a custom item retrieval logic and returns the map of items that
// is associated with the key.
// It returns nil if the item is not found/valid.
// It also ensures that only one execution of the wrapped Loader's LoadAll
// method is in-flight for a given key at a time.
func (l *SuppressedLoader[K, V]) LoadAll(ctx context.Context, c cache.Store[K, V], key K) (map[K]V, error) {
	// there should be a better/generic way to create a
	// singleflight Group's key. It's possible that a generic
	// singleflight.Group will be introduced with/in go1.19+
	strKey := defaultKeyEncoder(key)

	// the error can be discarded since the singleflight.Group
	// itself does not return any of its errors, it returns
	// the error that we return ourselves in the func below, which
	// is also nil
	res, err, _ := l.group.Do(strKey, func() (interface{}, error) {
		v, err := l.loader.LoadAll(ctx, c, key)
		if err != nil {
			return nil, err
		}

		return v, nil
	})
	if err != nil {
		return nil, err
	}

	if res == nil {
		return nil, nil
	}

	value, ok := res.(map[K]V)
	if !ok {
		return nil, fmt.Errorf("invalid type %T, expected %T", res, value)
	}

	return value, nil
}

// NewSuppressedLoader creates a new instance of suppressed loader.
func NewSuppressedLoader[K comparable, V any](loader cache.Loader[K, V]) cache.Loader[K, V] {
	return &SuppressedLoader[K, V]{
		loader: loader,
		group:  &singleflight.Group{},
	}
}
