package cachelocal

import (
	"time"

	"github.com/trinhdaiphuc/go-kit/cache"
)

type Option[K comparable, V any] func(*Options[K, V])

type Options[K comparable, V any] struct {
	Loader          cache.Loader[K, V]
	TTL             time.Duration
	CleanUpInterval time.Duration
}

func WithLoader[K comparable, V any](loader cache.Loader[K, V]) Option[K, V] {
	return func(o *Options[K, V]) {
		if loader != nil {
			o.Loader = loader
		}
	}
}

func WithTTL[K comparable, V any](ttl time.Duration) Option[K, V] {
	return func(o *Options[K, V]) {
		o.TTL = ttl
	}
}

func WithCleanUpInterval[K comparable, V any](interval time.Duration) Option[K, V] {
	return func(o *Options[K, V]) {
		o.CleanUpInterval = interval
	}
}

func defaultOption[K comparable, V any]() *Options[K, V] {
	return &Options[K, V]{
		Loader:          nil,
		TTL:             5 * time.Minute,
		CleanUpInterval: time.Hour,
	}
}
