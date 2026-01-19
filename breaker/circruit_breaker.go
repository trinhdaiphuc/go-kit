package breaker

import (
	"time"

	"github.com/sony/gobreaker/v2"

	"github.com/trinhdaiphuc/go-kit/metrics"
)

const (
	defaultCircuitBreakerName        = "breaker_default"
	defaultCircuitBreakerMaxRequests = 5
	defaultCircuitBreakerInterval    = 60 * time.Second
	defaultCircuitBreakerTimeout     = 60 * time.Second
)

// CircuitBreaker is a unified interface for both normal and distributed circuit breakers
type CircuitBreaker[T any] interface {
	// Execute runs the given request if the CircuitBreaker accepts it.
	// Execute returns an error instantly if the CircuitBreaker rejects the request.
	Execute(req func() (T, error)) (T, error)
	// Name returns the name of the CircuitBreaker.
	Name() string
	// State returns the current state of the CircuitBreaker.
	State() gobreaker.State
	// Counts returns internal counters
	Counts() gobreaker.Counts
}

// circuitBreakerWrapper wraps the normal circuit breaker
type circuitBreakerWrapper[T any] struct {
	cb *gobreaker.CircuitBreaker[T]
}

func (w *circuitBreakerWrapper[T]) Execute(req func() (T, error)) (T, error) {
	return w.cb.Execute(req)
}

func (w *circuitBreakerWrapper[T]) Name() string {
	return w.cb.Name()
}

func (w *circuitBreakerWrapper[T]) State() gobreaker.State {
	return w.cb.State()
}

func (w *circuitBreakerWrapper[T]) Counts() gobreaker.Counts {
	return w.cb.Counts()
}

// distributedCircuitBreakerWrapper wraps the distributed circuit breaker
type distributedCircuitBreakerWrapper[T any] struct {
	cb *gobreaker.DistributedCircuitBreaker[T]
}

func (w *distributedCircuitBreakerWrapper[T]) Execute(req func() (T, error)) (T, error) {
	return w.cb.Execute(req)
}

func (w *distributedCircuitBreakerWrapper[T]) Name() string {
	return w.cb.Name()
}

func (w *distributedCircuitBreakerWrapper[T]) State() gobreaker.State {
	state, _ := w.cb.State()
	return state
}

func (w *distributedCircuitBreakerWrapper[T]) Counts() gobreaker.Counts {
	return w.cb.Counts()
}

// NewCircuitBreaker creates a new CircuitBreaker.
// If a store is provided via WithStore option, it creates a distributed circuit breaker.
// Otherwise, it creates a normal circuit breaker.
func NewCircuitBreaker[T any](opts ...CircuitBreakerOptions) (CircuitBreaker[T], error) {
	options := newDefaultOption()
	for _, opt := range opts {
		opt(options)
	}

	if options.EnableMetric {
		options.OnStateChange = metrics.WrapOnStateChange(options.OnStateChange)
		options.IsSuccessful = metrics.WrapIsSuccessful(options.Name, options.IsSuccessful)
	}

	settings := gobreaker.Settings{
		Name:          options.Name,
		MaxRequests:   options.MaxRequests,
		Interval:      options.Interval,
		Timeout:       options.Timeout,
		ReadyToTrip:   options.ReadyToTrip,
		OnStateChange: options.OnStateChange,
		IsSuccessful:  options.IsSuccessful,
	}

	// If store is provided, create distributed circuit breaker
	if options.Store != nil {
		dcb, err := gobreaker.NewDistributedCircuitBreaker[T](options.Store, settings)
		if err != nil {
			return nil, err
		}
		return &distributedCircuitBreakerWrapper[T]{cb: dcb}, nil
	}

	// Otherwise, create normal circuit breaker
	return &circuitBreakerWrapper[T]{cb: gobreaker.NewCircuitBreaker[T](settings)}, nil
}

func newDefaultOption() *breakerOptions {
	return &breakerOptions{
		Name:        defaultCircuitBreakerName,
		MaxRequests: defaultCircuitBreakerMaxRequests,
		Interval:    defaultCircuitBreakerInterval,
		Timeout:     defaultCircuitBreakerTimeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return counts.Requests >= 3 && failureRatio >= 0.6
		},
		OnStateChange: nil,
		IsSuccessful:  nil,
	}
}

type CircuitBreakerOptions func(*breakerOptions)

type breakerOptions struct {
	Name          string
	MaxRequests   uint32
	Interval      time.Duration
	Timeout       time.Duration
	ReadyToTrip   metrics.ReadyToTripFunc
	OnStateChange metrics.OnStateChangeFunc
	IsSuccessful  metrics.IsSuccessfulFunc
	EnableMetric  bool
	Store         gobreaker.SharedDataStore
}

func WithCircuitBreakerName(name string) CircuitBreakerOptions {
	return func(opts *breakerOptions) {
		opts.Name = name
	}
}

func WithCircuitBreakerMaxRequests(maxRequests uint32) CircuitBreakerOptions {
	return func(opts *breakerOptions) {
		opts.MaxRequests = maxRequests
	}
}

func WithCircuitBreakerInterval(interval time.Duration) CircuitBreakerOptions {
	return func(opts *breakerOptions) {
		if interval <= 0 {
			interval = defaultCircuitBreakerInterval
		}
		opts.Interval = interval
	}
}

func WithCircuitBreakerTimeout(timeout time.Duration) CircuitBreakerOptions {
	return func(opts *breakerOptions) {
		if timeout <= 0 {
			timeout = defaultCircuitBreakerTimeout
		}
		opts.Timeout = timeout
	}
}

func WithCircuitBreakerReadyToTrip(readyToTrip metrics.ReadyToTripFunc) CircuitBreakerOptions {
	return func(opts *breakerOptions) {
		opts.ReadyToTrip = readyToTrip
	}
}

func WithCircuitBreakerOnStateChange(onStateChange metrics.OnStateChangeFunc) CircuitBreakerOptions {
	return func(opts *breakerOptions) {
		opts.OnStateChange = onStateChange
	}
}

func WithCircuitBreakerIsSuccessful(isSuccessful metrics.IsSuccessfulFunc) CircuitBreakerOptions {
	return func(opts *breakerOptions) {
		opts.IsSuccessful = isSuccessful
	}
}

func WithCircuitBreakerEnableMetric(enable bool) CircuitBreakerOptions {
	return func(opts *breakerOptions) {
		opts.EnableMetric = enable
	}
}

// WithStore sets the shared data store for distributed circuit breaker.
// If store is provided, NewCircuitBreaker() will create a distributed circuit breaker.
func WithStore(store gobreaker.SharedDataStore) CircuitBreakerOptions {
	return func(opts *breakerOptions) {
		opts.Store = store
	}
}
