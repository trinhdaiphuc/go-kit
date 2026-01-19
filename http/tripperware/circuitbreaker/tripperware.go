package circuitbreaker

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/trinhdaiphuc/go-kit/breaker"
	httptripperware "github.com/trinhdaiphuc/go-kit/http/tripperware"
)

// httpResponse wraps the HTTP response for circuit breaker
type httpResponse struct {
	body       []byte
	statusCode int
	header     http.Header
}

// Tripperware returns a tripperware that wraps requests with a circuit breaker.
// If the circuit breaker is open, requests will fail fast without making the actual HTTP call.
func Tripperware(opts ...breaker.CircuitBreakerOptions) httptripperware.Tripperware {
	// Add default HTTP-specific IsSuccessful function
	options := []breaker.CircuitBreakerOptions{
		breaker.WithCircuitBreakerIsSuccessful(defaultHTTPIsSuccessful),
	}
	options = append(options, opts...)

	cb, err := breaker.NewCircuitBreaker[*http.Response](options...)
	if err != nil {
		panic(err)
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return httptripperware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			resp, err := cb.Execute(func() (*http.Response, error) {
				return next.RoundTrip(req)
			})
			return resp, err
		})
	}
}

// defaultHTTPIsSuccessful determines if an HTTP response should be considered successful
// for the circuit breaker. Server errors (5xx) and timeouts are considered failures.
func defaultHTTPIsSuccessful(err error) bool {
	if err == nil {
		return true
	}

	// Timeout errors should trip the circuit breaker
	if os.IsTimeout(err) || errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return false
	}

	return true
}

// WithStatusCodeCheck returns an IsSuccessful function that also checks HTTP status codes.
// Status codes >= 500 are considered failures.
func WithStatusCodeCheck(next func(err error) bool) func(err error) bool {
	return func(err error) bool {
		if next != nil && !next(err) {
			return false
		}
		return true
	}
}
