package metrics

import (
	prom "github.com/prometheus/client_golang/prometheus"
)

// Monitor represents a collection of monitor to be registered on a
// Prometheus monitor registry for a http server.
type Monitor struct {
	serviceName           string
	requestRates          *prom.CounterVec
	durationSeconds       *prom.HistogramVec
	clientRequestRates    *prom.CounterVec
	clientDurationSeconds *prom.HistogramVec
	circuitBreakerState   *prom.GaugeVec
	requestCounter        *prom.CounterVec
	successCounter        *prom.CounterVec
	failureCounter        *prom.CounterVec
}

const (
	defaultServiceName = "default_service"
)

const (
	grpcLabelMethod     = "gRPC"
	producerLabelMethod = "producer"
	consumerLabelMethod = "consumer"
	cacheLabelMethod    = "cache"
	databaseLabelMethod = "database"
)

const (
	ServerCall = "server"
	ClientCall = "client"
)

var (
	metricLabels   = []string{"service_name", "method", "endpoint", "status_code", "http_status_code"}
	defaultBuckets = []float64{.005, .01, .02, .03, .05, .1, .2, .3, .5, 1, 2, 3, 5, 10}
	monitor        *Monitor
)

// NewServerMonitor returns a new ServerMetrics object.
func NewServerMonitor(serviceName string) *Monitor {
	if serviceName == "" {
		serviceName = defaultServiceName
	}

	monitor = &Monitor{
		serviceName: serviceName,
		requestRates: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "request_rates",
				Help: "Total number of request hit to the server.",
			},
			metricLabels,
		),
		durationSeconds: prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    "request_duration_seconds",
				Help:    "Histogram of response durationSeconds (seconds) of request that had been handled by the server.",
				Buckets: defaultBuckets,
			},
			metricLabels,
		),
		clientRequestRates: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "client_request_rates",
				Help: "Total number of request hit to the client.",
			},
			metricLabels,
		),
		clientDurationSeconds: prom.NewHistogramVec(
			prom.HistogramOpts{
				Name:    "client_request_duration_seconds",
				Help:    "Histogram of response durationSeconds (seconds) of request that had been handled by the client.",
				Buckets: defaultBuckets,
			},
			metricLabels,
		),
		circuitBreakerState: prom.NewGaugeVec(
			prom.GaugeOpts{
				Name: "breaker_state",
				Help: "The states of the circuit breaker. 0=Not Active, 1=Active. state=['open','half-open','closed']",
			},
			[]string{"service_name", "name", "state"},
		),
		requestCounter: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "breaker_requests_total",
				Help: "Total number of requests executed through the circuit breaker",
			},
			[]string{"service_name", "name"},
		),
		successCounter: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "breaker_success_total",
				Help: "Total number of successful requests",
			},
			[]string{"service_name", "name"},
		),
		failureCounter: prom.NewCounterVec(
			prom.CounterOpts{
				Name: "breaker_failure_total",
				Help: "Total number of failed requests",
			},
			[]string{"service_name", "name"},
		),
	}
	prom.MustRegister(
		monitor.requestRates,
		monitor.durationSeconds,
		monitor.circuitBreakerState,
		monitor.requestCounter,
		monitor.successCounter,
		monitor.failureCounter,
	)
	return monitor
}

func doneHandleRequest(callType, method, endpoint, statusCode, httpStatusCode string, observeTime float64) {
	if callType == ServerCall {
		monitor.requestRates.WithLabelValues(monitor.serviceName, method, endpoint, statusCode, httpStatusCode).Inc()
		monitor.durationSeconds.WithLabelValues(monitor.serviceName, method, endpoint, statusCode, httpStatusCode).Observe(observeTime)
	} else {
		monitor.clientRequestRates.WithLabelValues(monitor.serviceName, method, endpoint, statusCode, httpStatusCode).Inc()
		monitor.clientDurationSeconds.WithLabelValues(monitor.serviceName, method, endpoint, statusCode, httpStatusCode).Observe(observeTime)
	}
}
