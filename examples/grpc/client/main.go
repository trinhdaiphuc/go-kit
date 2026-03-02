package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/sony/gobreaker/v2"
	"github.com/trinhdaiphuc/go-kit/breaker"
	grpcclient "github.com/trinhdaiphuc/go-kit/grpc/client"
	grpcinterceptor "github.com/trinhdaiphuc/go-kit/grpc/interceptor"
	"github.com/trinhdaiphuc/go-kit/log"
	"github.com/trinhdaiphuc/go-kit/metrics"
	"github.com/trinhdaiphuc/go-kit/tracing"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
)

func main() {
	_, cleanupTracing, err := tracing.TracerProvider("grpc-client", "1.0.0", &tracing.OtelExporter{
		Jaeger: &tracing.Jaeger{
			AgentHost: "localhost",
			AgentPort: "6831",
		},
	})
	if err != nil {
		log.Bg().Fatal("Failed to initialize tracing", zap.Error(err))
	}
	defer cleanupTracing()

	metrics.NewServerMonitor("grpc-client")

	conn, err := grpcclient.NewClientConn("localhost:9090",
		grpcclient.WithUnaryClientInterceptors(
			grpcinterceptor.CircuitBreakerUnaryServerInterceptor(
				breaker.WithCircuitBreakerInterval(60*time.Second),
				breaker.WithCircuitBreakerMaxRequests(10),
				breaker.WithCircuitBreakerName("session_service"),
				breaker.WithCircuitBreakerTimeout(60*time.Second),
				breaker.WithCircuitBreakerEnableMetric(true),
				breaker.WithCircuitBreakerOnStateChange(func(name string, from gobreaker.State, to gobreaker.State) {
					log.Bg().Info("Circuit breaker state changed", zap.String("name", name), zap.String("from", from.String()), zap.String("to", to.String()))
				}),
				// breaker.WithCircuitBreakerIsSuccessful(func(err error) bool {
				// 	return err == nil
				// }),
				breaker.WithCircuitBreakerReadyToTrip(func(counts gobreaker.Counts) bool {
					failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
					isReadyToTrip := counts.Requests >= 3 && failureRatio >= 0.6
					if isReadyToTrip {
						log.Bg().Info("Circuit breaker ready to trip. Fail to close state", zap.Float64("failureRatio", failureRatio), zap.Uint32("requests", counts.Requests))
					}
					return isReadyToTrip
				}),
			),
			// grpcinterceptor.UnaryClientLoggerInterceptor(),
		),
		grpcclient.WithMonitoring(false),
	)
	if err != nil {
		log.Bg().Fatal("Failed to create client connection", zap.Error(err))
	}
	cli := pb.NewGreeterClient(conn)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	http.Handle("/metrics", promhttp.Handler())
	go func() {
		if err = http.ListenAndServe("127.0.0.1:8081", nil); err != nil {
			panic(err)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			log.Bg().Info("Client stopped")
			return
		default:
			ctx, span := tracing.CreateSpan(context.Background(), "grpc-client", trace.WithSpanKind(trace.SpanKindClient))
			_, err := cli.SayHello(ctx, &pb.HelloRequest{Name: "123"})
			if err != nil {
				// log.For(ctx).Error("Failed to get session", zap.Error(err))
			}
			// log.For(ctx).Info("Get session success", zap.Reflect("session", session))
			time.Sleep(50 * time.Millisecond)
			span.End()
		}
	}
}
