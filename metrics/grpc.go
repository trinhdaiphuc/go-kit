package metrics

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
)

func GrpcUnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		var httpStatusCode = http.StatusOK

		start := time.Now()
		resp, err := handler(ctx, req)
		elapsedTime := time.Since(start).Seconds()
		if err != nil {
			httpStatusCode = ParseErr(err)
		}

		httpStatusCodeStr := strconv.Itoa(httpStatusCode)
		method := grpcMethod(splitMethodName(info.FullMethod))
		doneHandleRequest(InboundCall, grpcLabelMethod, method, httpStatusCodeStr, elapsedTime)

		return resp, err
	}
}

func GrpcUnaryClientInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, fullMethod string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
		var httpStatusCode = http.StatusOK

		start := time.Now()
		err := invoker(ctx, fullMethod, req, reply, cc, callOpts...)
		elapsedTime := time.Since(start).Seconds()
		if err != nil {
			httpStatusCode = ParseErr(err)
		}

		// Current monitor doesn't use error code, so this acts as a placeholder for future use.
		httpStatusCodeStr := strconv.Itoa(httpStatusCode)
		method := grpcMethod(splitMethodName(fullMethod))
		doneHandleRequest(OutboundCall, grpcLabelMethod, method, httpStatusCodeStr, elapsedTime)

		return err
	}
}

func ParseErr(grpcErr error) int {
	err, ok := status.FromError(grpcErr)
	if !ok {
		return http.StatusInternalServerError
	}

	return runtime.HTTPStatusFromCode(err.Code())
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
