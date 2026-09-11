package cacheloader

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/trinhdaiphuc/go-kit/cache"
)

type blockingLoader struct {
	release chan struct{}
}

func (l *blockingLoader) Load(ctx context.Context, c cache.Store[string, *Data], key string) (*Data, error) {
	<-l.release
	return &Data{Name: "late"}, nil
}

func (l *blockingLoader) LoadAll(ctx context.Context, c cache.Store[string, *Data], key string) (map[string]*Data, error) {
	<-l.release
	return nil, nil
}

func (l *blockingLoader) BulkLoad(ctx context.Context, c cache.Store[string, *Data], keys []string) (map[string]*Data, error) {
	<-l.release
	return nil, nil
}

func TestSingleFlightLoader_LoadRespectsContext(t *testing.T) {
	loader := &blockingLoader{release: make(chan struct{})}
	defer close(loader.release)

	sf := NewSingleFlightLoader[string, *Data](loader)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	done := make(chan error, 1)
	go func() {
		_, err := sf.Load(ctx, nil, "key")
		done <- err
	}()

	select {
	case err := <-done:
		assert.ErrorIs(t, err, context.DeadlineExceeded)
	case <-time.After(time.Second):
		t.Fatal("Load ignored the context deadline")
	}
}
