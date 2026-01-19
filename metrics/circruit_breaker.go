package metrics

import (
	"github.com/sony/gobreaker/v2"
)

type OnStateChangeFunc func(name string, from gobreaker.State, to gobreaker.State)
type ReadyToTripFunc func(counts gobreaker.Counts) bool
type IsSuccessfulFunc func(err error) bool

func WrapOnStateChange(changeFunc OnStateChangeFunc) OnStateChangeFunc {
	return func(name string, from gobreaker.State, to gobreaker.State) {
		if changeFunc != nil {
			changeFunc(name, from, to)
		}
		monitor.circuitBreakerState.WithLabelValues(monitor.serviceName, name, to.String()).Set(float64(to))
	}
}

func WrapIsSuccessful(name string, isSuccessfulFunc IsSuccessfulFunc) IsSuccessfulFunc {
	return func(err error) bool {
		isSuccessful := err == nil
		if isSuccessfulFunc != nil {
			isSuccessful = isSuccessfulFunc(err)
		}
		monitor.requestCounter.WithLabelValues(monitor.serviceName, name).Inc()
		if isSuccessful {
			monitor.successCounter.WithLabelValues(monitor.serviceName, name).Inc()
			return true
		}
		monitor.failureCounter.WithLabelValues(monitor.serviceName, name).Inc()
		return false
	}
}
