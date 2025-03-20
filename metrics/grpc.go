package metrics

import (
	"context"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func GrpcUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		resp, err := handler(ctx, req)
		elapsedTime := time.Since(start).Seconds()

		method := grpcMethod(splitMethodName(info.FullMethod))
		doneGRPCHandleRequest(InboundCall, method, status.Code(err).String(), elapsedTime)

		return resp, err
	}
}

func GrpcUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, fullMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
		start := time.Now()
		err := invoker(ctx, fullMethod, req, reply, cc, callOpts...)
		elapsedTime := time.Since(start).Seconds()

		// Current monitor doesn't use error code, so this acts as a placeholder for future use.
		method := grpcMethod(splitMethodName(fullMethod))
		doneGRPCHandleRequest(OutboundCall, method, status.Code(err).String(), elapsedTime)

		return err
	}
}

func splitMethodName(fullMethodName string) (string, string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if i := strings.Index(fullMethodName, "/"); i >= 0 {
		return fullMethodName[:i], fullMethodName[i+1:]
	}
	return "unknown", "unknown"
}

func grpcMethod(serviceName, methodName string) string {
	return methodName + " (" + serviceName + ")"
}
