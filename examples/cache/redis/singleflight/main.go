package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/cache"
	cacheloader "github.com/trinhdaiphuc/go-kit/cache/loader"
	cacheredis "github.com/trinhdaiphuc/go-kit/cache/redis"
	"github.com/trinhdaiphuc/go-kit/log"
)

type User struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Birthday time.Time `json:"birthday"`
}

func userKeyFunc(key string) string {
	return "user:" + key
}

const (
	ttl = 5 * time.Second
)

func main() {
	cfg := &cacheredis.Config{
		Addresses: "localhost:6379",
		Prefix:    "example",
	}
	redisCli, close, err := cacheredis.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	defer close()

	redisCache := cacheredis.NewRedisCache[string, *User](
		redisCli,
		cacheredis.WithPrefix[string, *User]("example"),
		cacheredis.WithLoader[string, *User](
			cacheloader.NewSuppressedLoader[string, *User](&loader{}),
		),
	)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	for i := 0; i < 10; i++ {
		go func() {
			for {
				select {
				case <-ctx.Done():
					return
				default:
					_, err := redisCache.Get(ctx, "user")
					if err != nil {
						log.Bg().Error("Get user failed", zap.Error(err))
					}
				}
			}
		}()
	}

	<-ctx.Done()
}

type loader struct{}

func (l *loader) Load(ctx context.Context, c cache.Store[string, *User], key string) (*User, error) {
	user := &User{
		ID:       1,
		Name:     "John Doe",
		Birthday: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	time.Sleep(50 * time.Millisecond) // simulate slow loading

	log.Bg().Info("Load user", zap.String("key", key))
	err := c.Set(ctx, key, user, ttl)
	if err != nil {
		return nil, err
	}
	return user, nil
}

func (l *loader) LoadAll(ctx context.Context, c cache.Store[string, *User], key string) (map[string]*User, error) {
	users := map[string]*User{
		"user1": {
			ID:       1,
			Name:     "John Doe",
			Birthday: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		},
		"user2": {
			ID:       2,
			Name:     "Jenifer",
			Birthday: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
		},
	}

	keyVals := make([]cache.KeyVal[string, *User], 0, len(users))
	for key, user := range users {
		keyVals = append(keyVals, cache.KeyVal[string, *User]{Key: key, Value: user})
	}

	err := c.HSet(ctx, key, keyVals...)
	if err != nil {
		return nil, err
	}

	return users, nil
}
