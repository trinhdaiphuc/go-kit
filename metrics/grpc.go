package metrics

import (
	"context"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type HandleMethod func(fullMethod string, req any) string

func GrpcUnaryServerInterceptor(handleMethod ...HandleMethod) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req any, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (any, error) {
		var (
			httpStatusCode = http.StatusOK
			statusCode     = codes.OK
		)

		start := time.Now()
		resp, err := handler(ctx, req)
		elapsedTime := time.Since(start).Seconds()
		if err != nil {
			httpStatusCode = ParseErr(err)
			statusCode = status.Code(err)
		}

		httpStatusCodeStr := strconv.Itoa(httpStatusCode)
		method := GrpcMethod(SplitGRPCMethodName(info.FullMethod))
		if len(handleMethod) > 0 {
			method = handleMethod[0](info.FullMethod, req)
		}

		doneHandleRequest(ServerCall, grpcLabelMethod, method, statusCode.String(), httpStatusCodeStr, elapsedTime)

		return resp, err
	}
}

func GrpcUnaryClientInterceptor(handleMethod ...HandleMethod) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, fullMethod string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, callOpts ...grpc.CallOption) error {
		var (
			httpStatusCode = http.StatusOK
			statusCode     = codes.OK
		)

		start := time.Now()
		err := invoker(ctx, fullMethod, req, reply, cc, callOpts...)
		elapsedTime := time.Since(start).Seconds()
		if err != nil {
			httpStatusCode = ParseErr(err)
			statusCode = status.Code(err)
		}

		// Current monitor doesn't use error code, so this acts as a placeholder for future use.
		httpStatusCodeStr := strconv.Itoa(httpStatusCode)
		method := GrpcMethod(SplitGRPCMethodName(fullMethod))
		if len(handleMethod) > 0 {
			method = handleMethod[0](fullMethod, req)
		}

		doneHandleRequest(ClientCall, grpcLabelMethod, method, statusCode.String(), httpStatusCodeStr, elapsedTime)

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

func SplitGRPCMethodName(fullMethodName string) (serviceName, methodName string) {
	fullMethodName = strings.TrimPrefix(fullMethodName, "/") // remove leading slash
	if before, after, ok := strings.Cut(fullMethodName, "/"); ok {
		return before, after
	}
	return "unknown", "unknown"
}

func GrpcMethod(serviceName, methodName string) string {
	return methodName + " (" + serviceName + ")"
}
