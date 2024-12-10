package main

import (
	"context"
	"os"
	"os/signal"
	"time"

	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/cache"
	cacheloader "github.com/trinhdaiphuc/go-kit/cache/loader"
	cachelocal "github.com/trinhdaiphuc/go-kit/cache/local"
	"github.com/trinhdaiphuc/go-kit/log"
)

type User struct {
	ID       int       `json:"id"`
	Name     string    `json:"name"`
	Birthday time.Time `json:"birthday"`
}

func main() {
	cli := cachelocal.NewClient[int, *User](
		cachelocal.WithTTL[int, *User](5*time.Second),
		cachelocal.WithLoader[int, *User](cacheloader.NewSuppressedLoader(&loader{})),
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
					_, err := cli.Get(ctx, 1)
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

func (l *loader) Load(ctx context.Context, c cache.Store[int, *User], key int) (*User, error) {
	user := &User{
		ID:       1,
		Name:     "John Doe",
		Birthday: time.Date(1990, 1, 1, 0, 0, 0, 0, time.UTC),
	}

	time.Sleep(50 * time.Millisecond) // simulate slow loading

	log.Bg().Info("Load user", zap.Int("key", key))
	return user, nil
}

func (l *loader) LoadAll(ctx context.Context, c cache.Store[int, *User], key int) (map[int]*User, error) {
	return nil, nil
}
