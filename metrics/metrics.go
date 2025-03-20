package metrics

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

// Monitor represents a collection of monitor to be registered on a
// Prometheus monitor registry for a http server.
type Monitor struct {
	serviceName         string
	httpRequestRates    *prom.CounterVec
	httpDurationSeconds *prom.HistogramVec
	grpcRequestRates    *prom.CounterVec
	grpcDurationSeconds *prom.HistogramVec
}

const (
	DefaultServiceName   = "default_service"
	grpcLabelMethod      = "gRPC"
	producerLabelMethod  = "producer"
	consumerLabelMethod  = "consumer"
	cacheLabelMethod     = "cache"
	databaseLabelMethod  = "database"
	InboundCall          = "inbound"
	OutboundCall         = "outbound"
	consumeLabelMethod   = "consume"
	availableLabelMethod = "available"
)

var (
	metricLabels   = []string{"service_name", "call_type", "method", "endpoint", "http_status_code"}
	defaultBuckets = []float64{.005, .01, .02, .03, .05, .1, .2, .3, .5, 1, 2, 3, 5, 10}
	monitor        *Monitor
)

// NewServerMonitor returns a new ServerMetrics object.
func NewServerMonitor(serviceName string) *Monitor {
	if serviceName == "" {
		serviceName = DefaultServiceName
	}

	monitor = &Monitor{
		serviceName: serviceName,
		httpRequestRates: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "http_request_rates",
				Help: "Total number of request hit to the http server.",
			},
			[]string{"service_name", "call_type", "method", "endpoint", "status_code"},
		),
		httpDurationSeconds: prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    "http_request_duration_seconds",
				Help:    "Histogram of response httpDurationSeconds (seconds) of request that had been handled by the http server.",
				Buckets: defaultBuckets,
			},
			[]string{"service_name", "call_type", "method", "endpoint", "status_code"},
		),
		grpcRequestRates: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "grpc_request_rates",
				Help: "Total number of request hit to the gRPC server.",
			},
			[]string{"service_name", "call_type", "method", "status_code"},
		),
		grpcDurationSeconds: prom.NewHistogramVec(
			prom.HistogramOpts{
				Name: "grpc_request_duration_seconds",
				Help: "Histogram of response httpDurationSeconds (seconds) of request that had been handled by the gRPC server.",
			},
			[]string{"service_name", "call_type", "method", "status_code"},
		),
	}
	prom.MustRegister(
		monitor.httpRequestRates,
		monitor.httpDurationSeconds,
	)
	return monitor
}

func doneHTTPHandleRequest(callType, method, endpoint, statusCode string, observeTime float64) {
	monitor.httpRequestRates.WithLabelValues(monitor.serviceName, callType, method, endpoint, statusCode).Inc()
	monitor.httpDurationSeconds.WithLabelValues(monitor.serviceName, callType, method, endpoint, statusCode).Observe(observeTime)
}

func doneGRPCHandleRequest(callType, method, statusCode string, observeTime float64) {
	monitor.grpcRequestRates.WithLabelValues(monitor.serviceName, callType, method, statusCode).Inc()
	monitor.grpcDurationSeconds.WithLabelValues(monitor.serviceName, callType, method, statusCode).Observe(observeTime)
}
