package log

import (
	"context"

	"go.uber.org/zap"
	"google.golang.org/grpc/status"
)

type contextLogKey struct{}

type contextLogValue struct {
	apiName  string
	clientID string
}

var (
	contextLog = contextLogKey{}
)

func (c *contextLogValue) ToLoggerFields() []zap.Field {
	if c == nil {
		return nil
	}
	return []zap.Field{
		zap.String("api_name", c.apiName),
		zap.String("client_id", c.clientID),
	}
}

func NewCtxLogger(ctx context.Context, apiName, clientID string) context.Context {
	ctxLogger := &contextLogValue{apiName: apiName, clientID: clientID}
	return context.WithValue(ctx, contextLog, ctxLogger)
}

func GetContextLogData(ctx context.Context) []zap.Field {
	ctxData, ok := ctx.Value(contextLog).(*contextLogValue)
	if !ok {
		return nil
	}
	return ctxData.ToLoggerFields()
}

func NewLogger(ctx context.Context, fields ...zap.Field) Logger {
	return For(ctx, GetContextLogData).With(fields...)
}

func GRPCError(err error) zap.Field {
	return zap.Reflect("error_details", status.Convert(err).Details())
}
