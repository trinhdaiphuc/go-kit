package grpcclient

import (
	"crypto/tls"
	"time"

	grpcMiddleware "github.com/grpc-ecosystem/go-grpc-middleware"
	grpcRetry "github.com/grpc-ecosystem/go-grpc-middleware/retry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/backoff"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/keepalive"
	"google.golang.org/grpc/resolver"

	"github.com/trinhdaiphuc/go-kit/metrics"
)

func NewClientConn(address string, options ...Option) (*grpc.ClientConn, error) {
	var (
		opts []grpc.DialOption
		cfg  = defaultConfig()
	)

	for _, o := range options {
		o(cfg)
	}

	if cfg.TLSTransport {
		tlsCreds := credentials.NewTLS(&tls.Config{InsecureSkipVerify: true})
		opts = append(opts, grpc.WithTransportCredentials(tlsCreds))
	} else {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	optsRetry := []grpcRetry.CallOption{
		grpcRetry.WithBackoff(grpcRetry.BackoffExponential(50 * time.Millisecond)),
		grpcRetry.WithCodes(codes.Unavailable),
		grpcRetry.WithMax(cfg.RetryMaxTimes),
		grpcRetry.WithPerRetryTimeout(cfg.Timeout),
	}

	opts = append(
		opts,
		grpc.WithDefaultCallOptions(grpc.WaitForReady(true)),
		grpc.WithKeepaliveParams(
			keepalive.ClientParameters{
				Time:                5 * time.Minute,
				Timeout:             10 * time.Second,
				PermitWithoutStream: true,
			},
		),
		grpc.WithConnectParams(
			grpc.ConnectParams{
				Backoff: backoff.Config{
					BaseDelay:  time.Second,
					Multiplier: 2,
					Jitter:     0.2,
					MaxDelay:   120 * time.Second,
				},
				MinConnectTimeout: 15 * time.Second,
			},
		),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler()),
		grpc.WithDefaultServiceConfig(`{"loadBalancingPolicy":"round_robin"}`),
		grpc.WithDisableServiceConfig(),
	)

	sIntOpt := grpc.WithStreamInterceptor(
		grpcMiddleware.ChainStreamClient(
			append(
				[]grpc.StreamClientInterceptor{
					grpcRetry.StreamClientInterceptor(optsRetry...),
				}, cfg.StreamClientInterceptors...,
			)...,
		),
	)

	opts = append(opts, sIntOpt)

	unaryInterceptors := []grpc.UnaryClientInterceptor{
		grpcRetry.UnaryClientInterceptor(optsRetry...),
	}
	if cfg.MonitoringEnabled {
		unaryInterceptors = append(unaryInterceptors, metrics.GrpcUnaryClientInterceptor())
	}
	uIntOpt := grpc.WithUnaryInterceptor(grpcMiddleware.ChainUnaryClient(append(unaryInterceptors, cfg.UnaryClientInterceptors...)...))

	opts = append(opts, uIntOpt)

	if cfg.DNSResolver {
		opts = append(opts, grpc.WithResolvers(resolver.Get("dns")))
	}

	conn, err := grpc.NewClient(address, opts...)

	return conn, err
}
