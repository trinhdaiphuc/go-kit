package grpcinterceptor

import (
	"context"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"

	"github.com/trinhdaiphuc/go-kit/log"
)

type Fn func(c context.Context) []zapcore.Field

// LoggerMiddlewareConfig is config setting for Ginzap
type LoggerMiddlewareConfig struct {
	SkipPaths []string
	Context   Fn
}

// UnaryServerLoggerInterceptor for logging request and response gRPC unary server interceptor
func UnaryServerLoggerInterceptor(conf *LoggerMiddlewareConfig) grpc.UnaryServerInterceptor {
	skipPaths := make(map[string]bool, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skipPaths[path] = true
	}

	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (resp any, err error) {
		_, method := SplitGRPCMethodName(info.FullMethod)
		logger := log.NewLogger(ctx, zap.String("method", method))
		if conf.Context != nil {
			logger = logger.With(conf.Context(ctx)...)
		}

		ctx = log.NewCtxLogger(ctx, logger)

		// some evil middlewares modify this values
		if _, ok := skipPaths[info.FullMethod]; ok {
			return handler(ctx, req)
		}

		start := time.Now()

		resp, err = handler(ctx, req)

		fields := []zapcore.Field{
			zap.String("latency", time.Since(start).String()),
			zap.Reflect("request", req),
			zap.Reflect("response", resp),
		}

		message := "Finished unary call with"
		if err == nil {
			logger.Info(message, fields...)
			return
		}

		fields = append(fields, zap.Stringer("error_code", status.Code(err)), log.GRPCError(err), zap.Error(err))
		logger.Error(message, fields...)
		return
	}
}

func UnaryClientLoggerInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		serviceName, methodName := SplitGRPCMethodName(method)
		startTime := time.Now()

		err := invoker(ctx, method, req, reply, cc, opts...)

		log.For(ctx).Info("Call grpc service finish", zap.String("service", serviceName),
			zap.String("method", methodName), zap.String("latency", time.Since(startTime).String()))
		return err
	}
}
