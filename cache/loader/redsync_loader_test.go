package cacheloader

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"

	"github.com/trinhdaiphuc/go-kit/cache"
	redislock "github.com/trinhdaiphuc/go-kit/cache/redis/lock/mocks"
)

type redSyncLoaderMock struct {
	redLockMock *redislock.MockRedLock
	mutexMock   *redislock.MockLockMutex
}

type Data struct {
	Name  string `json:"name"`
	Value int    `json:"value"`
}

type loaderSuccess struct{}

func (l *loaderSuccess) Load(ctx context.Context, c cache.Store[string, *Data], key string) (*Data, error) {
	return &Data{Name: "John Doe", Value: 100}, nil
}

func (l *loaderSuccess) LoadAll(ctx context.Context, c cache.Store[string, *Data], key string) (map[string]*Data, error) {
	return map[string]*Data{
		"field1": {Name: "John Doe", Value: 100},
		"field2": {Name: "Jane Doe", Value: 200},
	}, nil
}

type loaderFailed struct{}

func (l *loaderFailed) Load(ctx context.Context, c cache.Store[string, *Data], key string) (*Data, error) {
	return nil, errors.New("failed")
}

func (l *loaderFailed) LoadAll(ctx context.Context, c cache.Store[string, *Data], key string) (map[string]*Data, error) {
	return nil, errors.New("failed")
}

func TestRedSyncLoader_Load(t *testing.T) {
	type args[K comparable, V any] struct {
		ctx    context.Context
		c      cache.Store[K, V]
		key    K
		loader cache.Loader[K, V]
	}
	type testCase[K comparable, V any] struct {
		name      string
		args      args[K, V]
		mock      func(redSyncLoaderMock)
		wantValue V
		wantErr   assert.ErrorAssertionFunc
	}
	tests := []testCase[string, *Data]{
		{
			name: "Test Load",
			args: args[string, *Data]{
				ctx:    context.Background(),
				c:      nil,
				key:    "key",
				loader: &loaderSuccess{},
			},
			mock: func(mock redSyncLoaderMock) {
				mock.redLockMock.EXPECT().GetLock("key", 1*time.Second).Return(mock.mutexMock)
				mock.mutexMock.EXPECT().TryLockContext(gomock.Any()).Return(nil)
				mock.mutexMock.EXPECT().Unlock().Return(true, nil)
			},
			wantValue: &Data{Name: "John Doe", Value: 100},
			wantErr:   assert.NoError,
		},
		{
			name: "Test Load with error",
			args: args[string, *Data]{
				ctx:    context.Background(),
				c:      nil,
				key:    "key",
				loader: &loaderFailed{},
			},
			mock: func(mock redSyncLoaderMock) {
				mock.redLockMock.EXPECT().GetLock("key", 1*time.Second).Return(mock.mutexMock)
				mock.mutexMock.EXPECT().TryLockContext(gomock.Any()).Return(nil)
				mock.mutexMock.EXPECT().Unlock().Return(true, nil)
			},
			wantValue: nil,
			wantErr:   assert.Error,
		},
		{
			name: "Test Load with error on TryLockContext",
			args: args[string, *Data]{
				ctx:    context.Background(),
				c:      nil,
				key:    "key",
				loader: &loaderFailed{},
			},
			mock: func(mock redSyncLoaderMock) {
				mock.redLockMock.EXPECT().GetLock("key", 1*time.Second).Return(mock.mutexMock)
				mock.mutexMock.EXPECT().TryLockContext(gomock.Any()).Return(fmt.Errorf("error"))
			},
			wantValue: nil,
			wantErr:   assert.Error,
		},
		{
			name: "Test Load with error on Unlock",
			args: args[string, *Data]{
				ctx:    context.Background(),
				c:      nil,
				key:    "key",
				loader: &loaderFailed{},
			},
			mock: func(mock redSyncLoaderMock) {
				mock.redLockMock.EXPECT().GetLock("key", 1*time.Second).Return(mock.mutexMock)
				mock.mutexMock.EXPECT().TryLockContext(gomock.Any()).Return(nil)
				mock.mutexMock.EXPECT().Unlock().Return(false, fmt.Errorf("error"))
			},
			wantValue: nil,
			wantErr:   assert.Error,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			redSyncLoader, redLockMock := newRedSyncLoaderMock(t, tt.args.loader)
			tt.mock(redLockMock)
			gotValue, err := redSyncLoader.Load(tt.args.ctx, tt.args.c, tt.args.key)
			if !tt.wantErr(t, err, fmt.Sprintf("Load(%v, %v, %v)", tt.args.ctx, tt.args.c, tt.args.key)) {
				return
			}
			assert.Equalf(t, tt.wantValue, gotValue, "Load(%v, %v, %v)", tt.args.ctx, tt.args.c, tt.args.key)
		})
	}
}

func keyFunc(key string) string {
	return key
}

func newRedSyncLoaderMock[K comparable, V any](t *testing.T, loader cache.Loader[K, V]) (*RedSyncLoader[K, V], redSyncLoaderMock) {
	ctrl := gomock.NewController(t)
	redLockMock := redislock.NewMockRedLock(ctrl)
	mutexMock := redislock.NewMockLockMutex(ctrl)
	return NewRedSyncLoader(redLockMock, loader, keyFunc, 1*time.Second), redSyncLoaderMock{
		redLockMock: redLockMock,
		mutexMock:   mutexMock,
	}
}
