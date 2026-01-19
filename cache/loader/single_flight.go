package cacheloader

import (
	"context"

	"golang.org/x/sync/singleflight"

	"github.com/trinhdaiphuc/go-kit/cache"
)

type SingleFlightLoader[K comparable, V any] struct {
	loader cache.Loader[K, V]
	singleflight.Group
}

func NewSingleFlightLoader[K comparable, V any](loader cache.Loader[K, V]) *SingleFlightLoader[K, V] {
	return &SingleFlightLoader[K, V]{
		loader: loader,
		Group:  singleflight.Group{},
	}
}

func (s *SingleFlightLoader[K, V]) Load(ctx context.Context, c cache.Store[K, V], key K) (value V, err error) {
	out, err, _ := s.Group.Do(defaultKeyEncoder(key), func() (interface{}, error) {
		return s.loader.Load(ctx, c, key)
	})
	if err != nil {
		var zero V
		return zero, err
	}
	return out.(V), nil
}

func (s *SingleFlightLoader[K, V]) LoadAll(ctx context.Context, c cache.Store[K, V], key K) (map[K]V, error) {
	out, err, _ := s.Group.Do(defaultKeyEncoder(key), func() (interface{}, error) {
		return s.loader.LoadAll(ctx, c, key)
	})
	if err != nil {
		return nil, err
	}

	return out.(map[K]V), nil
}

func (s *SingleFlightLoader[K, V]) BulkLoad(ctx context.Context, c cache.Store[K, V], keys []K) (map[K]V, error) {
	out, err, _ := s.Group.Do(defaultKeyEncoder(keys), func() (interface{}, error) {
		return s.loader.BulkLoad(ctx, c, keys)
	})
	if err != nil {
		return nil, err
	}

	return out.(map[K]V), nil
}
