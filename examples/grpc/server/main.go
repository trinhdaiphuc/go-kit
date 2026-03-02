package main

import (
	"context"
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	pb "google.golang.org/grpc/examples/helloworld/helloworld"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/metadata"

	grpcinterceptor "github.com/trinhdaiphuc/go-kit/grpc/interceptor"
	grpcserver "github.com/trinhdaiphuc/go-kit/grpc/server"
	"github.com/trinhdaiphuc/go-kit/log"
	"github.com/trinhdaiphuc/go-kit/metrics"
	"github.com/trinhdaiphuc/go-kit/tracing"
)

func main() {
	listen, err := net.Listen("tcp", "127.0.0.1:9090")
	if err != nil {
		panic(err)
	}

	_, cleanupTracing, err := tracing.TracerProvider("grpc-server", "1.0.0", &tracing.OtelExporter{
		Jaeger: &tracing.Jaeger{
			AgentHost: "localhost",
			AgentPort: "6831",
		},
	})
	if err != nil {
		log.Bg().Fatal("Failed to initialize tracing", zap.Error(err))
	}

	// Start gRPC server
	metrics.NewServerMonitor("test")
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			metrics.GrpcUnaryServerInterceptor(),
			grpcinterceptor.UnaryServerLoggerInterceptor(&grpcinterceptor.LoggerMiddlewareConfig{
				Context: clientIDContextLogField,
			}),
		),
	)

	grpc_health_v1.RegisterHealthServer(grpcServer, grpcserver.NewHealthController(&healthService{}))
	pb.RegisterGreeterServer(grpcServer, &server{})

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err = grpcServer.Serve(listen); err != nil {
			panic(err)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		if err = http.ListenAndServe("127.0.0.1:8080", nil); err != nil {
			panic(err)
		}
	}()

	log.Bg().Info("Server started")

	<-stop
	log.Bg().Info("Server stopped")
	grpcServer.GracefulStop()
	cleanupTracing()
}

// server is used to implement helloworld.GreeterServer.
type server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *server) SayHello(_ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.For(context.Background()).Info("Received: ", zap.String("name", in.GetName()))
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
}

type healthService struct {
}

func (h *healthService) Check(ctx context.Context) error {
	logger := log.For(ctx).With(zap.String("service", "health"))
	logger.Info("Start Health check")

	checkDb(logger)
	checkCache(ctx)
	return nil
}

func checkDb(logger log.Logger) {
	logger.Info("Checking db success")
}

func checkCache(ctx context.Context) {
	log.For(ctx).Info("Checking cache success")
}

var (
	_ grpcserver.Service = (*healthService)(nil)
)

func clientIDContextLogField(ctx context.Context) []zapcore.Field {
	header, ok := metadata.FromIncomingContext(ctx)
	if !ok {
		return nil
	}
	clientIDs := header.Get("client-id")
	if len(clientIDs) == 0 {
		return nil
	}
	return []zapcore.Field{zap.Strings("client_id", clientIDs)}
}
