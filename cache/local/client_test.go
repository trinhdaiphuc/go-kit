package cachelocal

import (
	"context"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
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
