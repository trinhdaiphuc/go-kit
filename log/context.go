package log

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type contextLogKey struct{}

var (
	contextLog = contextLogKey{}
)

func NewCtxLogger(ctx context.Context, logger Logger) context.Context {
	return context.WithValue(ctx, contextLog, logger)
}

func GetContextLogData(ctx context.Context) Logger {
	logger, ok := ctx.Value(contextLog).(Logger)
	if !ok {
		return nil
	}
	return logger
}

func NewLogger(ctx context.Context, fields ...zap.Field) Logger {
	return For(ctx).With(fields...)
}

func GRPCError(err error) zap.Field {
	return zap.Reflect("error_details", status.Convert(err).Details())
}
