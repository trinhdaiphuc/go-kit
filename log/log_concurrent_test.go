package log

import (
	"context"
	"sync"
	"testing"
)

// Fails under -race on the old `if instance == nil` fast path.
func TestConcurrentBgAndFor(t *testing.T) {
	var wg sync.WaitGroup
	for i := 0; i < 50; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); Bg().Info("bg") }()
		go func() { defer wg.Done(); For(context.Background()).Info("for") }()
	}
	wg.Wait()
}
