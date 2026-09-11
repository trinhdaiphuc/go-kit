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
	out, err := s.do(ctx, defaultKeyEncoder(key), func() (any, error) {
		return s.loader.Load(ctx, c, key)
	})
	if err != nil {
		var zero V
		return zero, err
	}
	return out.(V), nil
}

// do runs fn under the singleflight group but lets the caller give up when its
// context ends: Group.Do blocks until the leader returns, so one slow leader
// otherwise pins every follower past its own deadline.
func (s *SingleFlightLoader[K, V]) do(ctx context.Context, key string, fn func() (any, error)) (any, error) {
	ch := s.Group.DoChan(key, fn)
	select {
	case res := <-ch:
		return res.Val, res.Err
	case <-ctx.Done():
		s.Group.Forget(key)
		return nil, ctx.Err()
	}
}

func (s *SingleFlightLoader[K, V]) LoadAll(ctx context.Context, c cache.Store[K, V], key K) (map[K]V, error) {
	out, err := s.do(ctx, defaultKeyEncoder(key), func() (any, error) {
		return s.loader.LoadAll(ctx, c, key)
	})
	if err != nil {
		return nil, err
	}

	return out.(map[K]V), nil
}

func (s *SingleFlightLoader[K, V]) BulkLoad(ctx context.Context, c cache.Store[K, V], keys []K) (map[K]V, error) {
	out, err := s.do(ctx, defaultKeyEncoder(keys), func() (any, error) {
		return s.loader.BulkLoad(ctx, c, keys)
	})
	if err != nil {
		return nil, err
	}

	return out.(map[K]V), nil
}
