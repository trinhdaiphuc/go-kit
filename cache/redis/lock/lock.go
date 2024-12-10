package redislock

import (
	"context"
	"time"

	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/redis/go-redis/v9"
)

//go:generate mockgen -destination=./mocks/$GOFILE -source=$GOFILE -package=redislock
type LockMutex interface {
	TryLockContext(ctx context.Context) error
	Unlock() (bool, error)
}

type RedLock interface {
	GetLock(key string, expiry time.Duration) LockMutex
}

type redisRedLock struct {
	opts      *Options
	redisLock *redsync.Redsync
}

func NewRedLock(redisCli redis.UniversalClient, options ...Option) RedLock {
	pool := goredis.NewPool(redisCli)
	redisLock := redsync.New(pool)

	opts := newDefaultOption()
	for _, o := range options {
		o(opts)
	}
	return &redisRedLock{
		opts:      opts,
		redisLock: redisLock,
	}
}

func (r *redisRedLock) GetLock(key string, expiry time.Duration) LockMutex {
	return r.redisLock.NewMutex(r.buildKey(key), redsync.WithExpiry(expiry))
}

func (r *redisRedLock) buildKey(key string) string {
	return r.opts.prefix + ":" + key + ":" + r.opts.suffix
}
