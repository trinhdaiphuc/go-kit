// Copyright 2017 Michal Witkowski. All Rights Reserved.
// See LICENSE for licensing terms.

package tripperwareretry

import (
	"context"
	"errors"
	"net/http"
	"time"

	httptripperware "github.com/trinhdaiphuc/go-kit/http/tripperware"
)

// Tripperware is client side HTTP ware that retries the requests.
//
// Be default this retries safe and idempotent requests 3 times with a linear delay of 100ms. This behaviour can be
// customized using With* parameter options.
//
// Requests that have `http_retry.Enable` set on them will always be retried.
func Tripperware(opts ...Option) httptripperware.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		o := evaluateOptions(opts)
		return httptripperware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			// Short-circuit to avoid allocations.
			if !o.decider(req) {
				return next.RoundTrip(req)
			}
			if o.maxRetry == 0 {
				return next.RoundTrip(req)
			}

			var (
				err      error
				lastResp *http.Response
			)

			for attempt := uint(0); attempt < o.maxRetry; attempt++ {
				thisReq := req.WithContext(req.Context()) // make a copy.
				if err := waitRetryBackoff(attempt, req.Context(), o); err != nil {
					return nil, err // context errors from req.Context()
				}
				lastResp, err = next.RoundTrip(thisReq)
				if isContextError(err) {
					break // do not retry context errors
				}

				if err == nil && !isRetryable(lastResp, o) {
					break // do not retry responses that the discarder tells us we should not discard
				}
			}

			return lastResp, err
		})
	}
}

func waitRetryBackoff(attempt uint, parentCtx context.Context, opt *options) error {
	var waitTime time.Duration = 0
	if attempt > 0 {
		waitTime = opt.backoffFunc(attempt)
	}
	if waitTime > 0 {
		select {
		case <-parentCtx.Done():
			return parentCtx.Err()
		case <-time.After(waitTime):
		}
	}
	return nil
}

func isContextError(err error) bool {
	return errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled)
}

func isRetryable(response *http.Response, o *options) bool {
	if response == nil {
		return false
	}

	for _, code := range o.retryCodes {
		if response.StatusCode == code {
			return true
		}
	}
	return false
}
