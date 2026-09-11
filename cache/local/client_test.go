package cachelocal

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/trinhdaiphuc/go-kit/cache"
)

type data struct {
	Name string
}

func TestClient_CloseStopsCleanupGoroutine(t *testing.T) {
	before := runtime.NumGoroutine()

	c := NewClient[string, *data](WithCleanUpInterval[string, *data](10 * time.Millisecond))
	assert.NoError(t, c.Set(context.Background(), "key", &data{Name: "John Doe"}))
	c.Close()

	// The cleanup goroutine exits asynchronously; poll instead of sleeping once.
	deadline := time.Now().Add(2 * time.Second)
	for runtime.NumGoroutine() > before && time.Now().Before(deadline) {
		time.Sleep(10 * time.Millisecond)
	}

	assert.LessOrEqual(t, runtime.NumGoroutine(), before, "cleanUpExpired goroutine still running after Close")
}

func TestClient_CloseIsIdempotent(t *testing.T) {
	c := NewClient[string, *data]()

	assert.NotPanics(t, func() {
		c.Close()
		c.Close()
	})
}

func TestClient_UnsupportedOperations(t *testing.T) {
	c := NewClient[string, *data]()
	defer c.Close()

	ctx := context.Background()

	t.Run("Incr", func(t *testing.T) {
		_, err := c.Incr(ctx, "key", 1)
		assert.ErrorIs(t, err, cache.ErrorUnsupportedOperation)
	})

	t.Run("HSet", func(t *testing.T) {
		assert.ErrorIs(t, c.HSet(ctx, "key"), cache.ErrorUnsupportedOperation)
	})

	t.Run("HGet", func(t *testing.T) {
		_, err := c.HGet(ctx, "key", "field")
		assert.ErrorIs(t, err, cache.ErrorUnsupportedOperation)
	})

	t.Run("HGetAll", func(t *testing.T) {
		_, err := c.HGetAll(ctx, "key")
		assert.ErrorIs(t, err, cache.ErrorUnsupportedOperation)
	})

	t.Run("HDel", func(t *testing.T) {
		assert.ErrorIs(t, c.HDel(ctx, "key", "field"), cache.ErrorUnsupportedOperation)
	})
}
