package cacheredis

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/trinhdaiphuc/go-kit/cache"
)

// Set REDIS_ADDR to point at another instance.
func newIntegrationClient(t *testing.T) redis.UniversalClient {
	t.Helper()

	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}

	cli := redis.NewClient(&redis.Options{Addr: addr})
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	if err := cli.Ping(ctx).Err(); err != nil {
		t.Skipf("no redis at %s: %v", addr, err)
	}

	t.Cleanup(func() {
		_ = cli.Close()
	})

	return cli
}

type intLoader struct {
	all map[string]*Data
}

func (l *intLoader) Load(ctx context.Context, c cache.Store[string, *Data], key string) (*Data, error) {
	return nil, cache.ErrorKeyNotFound
}

func (l *intLoader) LoadAll(ctx context.Context, c cache.Store[string, *Data], key string) (map[string]*Data, error) {
	return l.all, nil
}

func (l *intLoader) BulkLoad(ctx context.Context, c cache.Store[string, *Data], keys []string) (map[string]*Data, error) {
	return nil, nil
}

func TestIntegration_BulkGetReturnsOriginalKeys(t *testing.T) {
	cli := newIntegrationClient(t)
	ctx := context.Background()
	prefix := "gokit-it-bulkget"
	t.Cleanup(func() { cli.Del(ctx, prefix+":a", prefix+":b") })

	repo := NewRedisCache[string, *Data](cli, WithPrefix[string, *Data](prefix), WithTTL[string, *Data](time.Minute))
	require.NoError(t, repo.Set(ctx, "a", &Data{Name: "A", Value: 1}))
	require.NoError(t, repo.Set(ctx, "b", &Data{Name: "B", Value: 2}))

	got, err := repo.BulkGet(ctx, []string{"a", "b"})
	require.NoError(t, err)
	assert.Equal(t, map[string]*Data{
		"a": {Name: "A", Value: 1},
		"b": {Name: "B", Value: 2},
	}, got, "result must be keyed by the caller's key, not the prefixed one")
}

func TestIntegration_HGetHydratesFromLoader(t *testing.T) {
	cli := newIntegrationClient(t)
	ctx := context.Background()
	prefix := "gokit-it-hget"
	t.Cleanup(func() { cli.Del(ctx, prefix+":hash") })
	cli.Del(ctx, prefix+":hash")

	loader := &intLoader{all: map[string]*Data{"field1": {Name: "John", Value: 100}}}
	repo := NewRedisCache[string, *Data](cli,
		WithPrefix[string, *Data](prefix),
		WithTTL[string, *Data](time.Minute),
		WithLoader[string, *Data](loader),
	)

	got, err := repo.HGet(ctx, "hash", "field1")
	require.NoError(t, err)
	assert.Equal(t, &Data{Name: "John", Value: 100}, got)

	// The hydrated hash must have been written back under the hash key.
	all, err := repo.HGetAll(ctx, "hash")
	require.NoError(t, err)
	assert.Equal(t, map[string]*Data{"field1": {Name: "John", Value: 100}}, all)
}

func TestIntegration_SetNXOnlyOnce(t *testing.T) {
	cli := newIntegrationClient(t)
	ctx := context.Background()
	prefix := "gokit-it-setnx"
	t.Cleanup(func() { cli.Del(ctx, prefix+":lock") })
	cli.Del(ctx, prefix+":lock")

	repo := NewRedisCache[string, *Data](cli, WithPrefix[string, *Data](prefix), WithTTL[string, *Data](time.Minute))

	ok, err := repo.SetNX(ctx, "lock", &Data{Name: "first"})
	require.NoError(t, err)
	assert.True(t, ok)

	ok, err = repo.SetNX(ctx, "lock", &Data{Name: "second"})
	require.NoError(t, err)
	assert.False(t, ok)
}
