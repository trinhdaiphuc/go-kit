package metrics

import (
	"context"
	"net"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
)

type redisHook struct{}

var (
	pingCmd = "ping"
)

func NewRedisHook() redis.Hook {
	return &redisHook{}
}

func (metric *redisHook) DialHook(next redis.DialHook) redis.DialHook {
	return func(ctx context.Context, network, addr string) (net.Conn, error) {
		return next(ctx, network, addr)
	}
}

func (metric *redisHook) ProcessHook(next redis.ProcessHook) redis.ProcessHook {
	return func(ctx context.Context, cmd redis.Cmder) error {
		if cmd.Name() == pingCmd { // ignore ping cmd
			return next(ctx, cmd)
		}

		startTime := time.Now()
		err := next(ctx, cmd)
		elapsedTime := time.Since(startTime).Seconds()

		statusCode := getRedisErrorCode(err)

		doneHandleRequest(OutboundCall, cacheLabelMethod, cmd.Name(), statusCode, elapsedTime)

		return err
	}
}

func (metric *redisHook) ProcessPipelineHook(next redis.ProcessPipelineHook) redis.ProcessPipelineHook {
	return func(ctx context.Context, cmds []redis.Cmder) error {
		startTime := time.Now()
		err := next(ctx, cmds)
		elapsedTime := time.Since(startTime).Seconds()

		statusCode := getRedisErrorCode(err)

		doneHandleRequest(OutboundCall, cacheLabelMethod, getCmdsName(cmds), statusCode, elapsedTime)

		return err
	}
}

func getRedisErrorCode(err error) string {
	var statusCode string
	switch err {
	case nil:
		// Success
		statusCode = "200"
	case redis.Nil:
		// Not found
		statusCode = "400"
	default:
		// Internal error
		statusCode = "500"
	}
	return statusCode
}

func getCmdsName(cmds []redis.Cmder) string {
	var cmdsStr []string
	for _, cmd := range cmds {
		cmdsStr = append(cmdsStr, cmd.Name())
	}
	return strings.Join(cmdsStr, " ")
}
