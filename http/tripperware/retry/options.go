// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package tripperwareretry

import (
	"net/http"
	"time"
)

var (
	defaultRetryableCodes = []int{500, 501, 502, 503, 504, 505, 506, 507, 508, 510, 511}
	defaultOptions        = &options{
		decider:     DefaultRetriableDecider,
		retryCodes:  defaultRetryableCodes,
		maxRetry:    3,
		backoffFunc: BackoffLinear(100 * time.Millisecond),
	}
)

type options struct {
	decider     RequestRetryDeciderFunc
	retryCodes  []int
	maxRetry    uint
	backoffFunc BackoffFunc
}

func evaluateOptions(opts []Option) *options {
	optCopy := &options{}
	*optCopy = *defaultOptions
	for _, o := range opts {
		o(optCopy)
	}
	return optCopy
}

type Option func(*options)

// RequestRetryDeciderFunc decides whether the given function is idempotent and safe or to retry.
type RequestRetryDeciderFunc func(req *http.Request) bool

// ResponseDiscarderFunc decides when to discard a response and retry the request again (on true).
type ResponseDiscarderFunc func(resp *http.Response) bool

// BackoffFunc denotes a family of functions that controll the backoff duration between call retries.
//
// They are called with an identifier of the attempt, and should return a time the system client should
// hold off for. If the time returned is longer than the `context.Context.Deadline` of the request
// the deadline of the request takes precedence and the wait will be interrupted before proceeding
// with the next iteration.
type BackoffFunc func(attempt uint) time.Duration

// WithMax sets the maximum number of retries on this call, or this interceptor.
func WithMax(maxRetries uint) Option {
	return func(o *options) {
		o.maxRetry = maxRetries
	}
}

// WithBackoff sets the `BackoffFunc `used to control time between retries.
func WithBackoff(bf BackoffFunc) Option {
	return func(o *options) {
		o.backoffFunc = bf
	}
}

// WithDecider is a function that allows users to customize the logic that decides whether a request is retriable.
func WithDecider(f RequestRetryDeciderFunc) Option {
	return func(o *options) {
		o.decider = f
	}
}

// WithRetryCodes sets the list of HTTP status codes that are retriable.
func WithRetryCodes(codes ...int) Option {
	return func(o *options) {
		o.retryCodes = append(o.retryCodes, codes...)
	}
}

// DefaultRetriableDecider is the default implementation that retries only indempotent and safe requests (GET, OPTION, HEAD).
//
// It is fairly conservative and heeds the of http://restcookbook.com/HTTP%20Methods/idempotency.
func DefaultRetriableDecider(req *http.Request) bool {
	if req.Method == http.MethodGet || req.Method == http.MethodOptions || req.Method == http.MethodHead {
		return true
	}
	return false
}
