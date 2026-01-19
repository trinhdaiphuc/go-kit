package grpcinterceptor

import (
	"context"
	"errors"
	"os"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/trinhdaiphuc/go-kit/breaker"
)

func CircuitBreakerUnaryServerInterceptor(opts ...breaker.CircuitBreakerOptions) grpc.UnaryClientInterceptor {
	options := []breaker.CircuitBreakerOptions{
		breaker.WithCircuitBreakerIsSuccessful(defaultGrpcClientIsSuccessful),
	}
	options = append(options, opts...)
	cb, err := breaker.NewCircuitBreaker[any](options...)
	if err != nil {
		panic(err)
	}
	return func(ctx context.Context, method string, req, reply any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		_, cbErr := cb.Execute(func() (any, error) {
			err := invoker(ctx, method, req, reply, cc, opts...)
			if err != nil {
				return nil, err
			}
			return nil, nil
		})
		return cbErr
	}
}

func defaultGrpcClientIsSuccessful(err error) bool {
	if err == nil {
		return true
	}

	code := status.Code(err)
	if os.IsTimeout(err) ||
		errors.Is(err, context.DeadlineExceeded) ||
		code == codes.ResourceExhausted ||
		code == codes.Unimplemented ||
		code == codes.Unavailable ||
		code == codes.DeadlineExceeded {

		return false
	}

	return true
}
