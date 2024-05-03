package metrics

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

// Monitor represents a collection of monitor to be registered on a
// Prometheus monitor registry for a http server.
type Monitor struct {
	serviceName     string
	requestRates    *prom.CounterVec
	durationSeconds *prom.HistogramVec
}

const (
	DefaultServiceName  = "default_service"
	grpcLabelMethod     = "gRPC"
	producerLabelMethod = "producer"
	consumerLabelMethod = "consumer"
	cacheLabelMethod    = "cache"
	databaseLabelMethod = "database"
	InboundCall         = "inbound"
	OutboundCall        = "outbound"
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
		requestRates: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "go_kit_request_rates",
				Help: "Total number of request hit to the server.",
			},
			metricLabels,
		),
		durationSeconds: prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    "go_kit_request_duration_seconds",
				Help:    "Histogram of response durationSeconds (seconds) of request that had been handled by the server.",
				Buckets: defaultBuckets,
			},
			metricLabels,
		),
	}
	prom.MustRegister(
		monitor.requestRates,
		monitor.durationSeconds,
	)
	return monitor
}

func doneHandleRequest(callType, method, endpoint, httpStatusCode string, observeTime float64) {
	monitor.requestRates.WithLabelValues(monitor.serviceName, callType, method, endpoint, httpStatusCode).Inc()
	monitor.durationSeconds.WithLabelValues(
		monitor.serviceName, callType, method, endpoint, httpStatusCode,
	).Observe(observeTime)
}
