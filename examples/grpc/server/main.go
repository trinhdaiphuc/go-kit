package main

import (
	"net"
	"net/http"
	"os"
	"os/signal"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	"github.com/trinhdaiphuc/go-kit/log"
	"github.com/trinhdaiphuc/go-kit/metrics"
	"github.com/trinhdaiphuc/go-kit/tracing"
)

func main() {
	listen, err := net.Listen("tcp", ":9090")
	if err != nil {
		panic(err)
	}

	_, _, err = tracing.TracerProvider("test", "1.0.0")
	if err != nil {
		panic(err)
	}

	// Start gRPC server
	metrics.NewServerMonitor("test")
	grpcServer := grpc.NewServer(
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.ChainUnaryInterceptor(
			metrics.GrpcUnaryServerInterceptor(),
		),
	)

	grpc_health_v1.RegisterHealthServer(grpcServer, health.NewServer())

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)

	go func() {
		if err = grpcServer.Serve(listen); err != nil {
			panic(err)
		}
	}()

	http.Handle("/metrics", promhttp.Handler())

	go func() {
		if err = http.ListenAndServe(":8080", nil); err != nil {
			panic(err)
		}
	}()

	log.Bg().Info("Server started")

	<-stop
	log.Bg().Info("Server stopped")
	grpcServer.GracefulStop()
}
