package tripperwareretry

import (
	"context"
	"errors"
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	httptripperware "github.com/trinhdaiphuc/go-kit/http/tripperware"
)

func TestTripperware(t *testing.T) {
	type args struct {
		opts         []Option
		request      *http.Request
		roundTripper http.RoundTripper
	}
	tests := []struct {
		name           string
		args           args
		wantRetryTimes int
		wantResponse   *http.Response
		wantErr        error
	}{
		{
			name: "Test Tripperware return success",
			args: args{
				opts: []Option{},
				request: &http.Request{
					Method: "GET",
				},
				roundTripper: httptripperware.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
					}, nil
				}),
			},
			wantRetryTimes: 1,
			wantResponse: &http.Response{
				StatusCode: 200,
			},
			wantErr: nil,
		},
		{
			name: "Test Tripperware return error",
			args: args{
				opts: []Option{
					WithBackoff(func(attempt uint) time.Duration {
						return 50 * time.Millisecond
					}),
					WithMax(0),
				},
				request: &http.Request{
					Method: "GET",
				},
				roundTripper: httptripperware.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
					return nil, errors.New("error")
				}),
			},
			wantRetryTimes: 1,
			wantResponse:   nil,
			wantErr:        errors.New("error"),
		},
		{
			name: "Test Tripperware not retry when request is not retryable",
			args: args{
				opts: []Option{
					WithDecider(func(r *http.Request) bool {
						return false
					}),
				},
				request: &http.Request{
					Method: "GET",
				},
				roundTripper: httptripperware.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
					return &http.Response{
						StatusCode: 200,
					}, nil
				}),
			},
			wantRetryTimes: 1,
			wantResponse: &http.Response{
				StatusCode: 200,
			},
			wantErr: nil,
		},
		{
			name: "Test Tripperware return error when wait time is greater than context deadline",
			args: args{
				opts: []Option{
					WithBackoff(func(attempt uint) time.Duration {
						return 100 * time.Millisecond
					}),
					WithMax(3),
				},
				request: func() *http.Request {
					ctx, _ := context.WithTimeout(context.Background(), 250*time.Millisecond)
					r := &http.Request{
						Method: "GET",
					}
					return r.WithContext(ctx)
				}(),
				roundTripper: httptripperware.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
					time.Sleep(100 * time.Millisecond)
					return &http.Response{
						StatusCode: 500,
					}, nil
				}),
			},
			wantRetryTimes: 2,
			wantResponse:   nil,
			wantErr:        context.DeadlineExceeded,
		},
		{
			name: "Test Tripperware return error when context is canceled",
			args: args{
				opts: []Option{
					WithBackoff(func(attempt uint) time.Duration {
						return 100 * time.Millisecond
					}),
					WithMax(2),
				},
				request: &http.Request{
					Method: "GET",
				},
				roundTripper: httptripperware.RoundTripperFunc(func(request *http.Request) (*http.Response, error) {
					time.Sleep(500 * time.Millisecond)
					return nil, context.DeadlineExceeded
				}),
			},
			wantRetryTimes: 1,
			wantResponse:   nil,
			wantErr:        context.DeadlineExceeded,
		},
	}
	request := &http.Request{
		Method: "GET",
	}
	request.WithContext(context.Background())
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			counter := counterRoundTripperWare{}
			tripper := Tripperware(tt.args.opts...)(counter.Tripper()(tt.args.roundTripper))
			resp, err := tripper.RoundTrip(tt.args.request)
			assert.Equalf(t, tt.wantErr, err, "Tripperware() error = %v, wantErr %v", err, tt.wantErr)
			assert.Equalf(t, tt.wantResponse, resp, "Tripperware() = %v, want %v", resp, tt.wantResponse)
			assert.Equalf(t, tt.wantRetryTimes, counter.count, "Tripperware() retry times = %v, want %v", counter.count, tt.wantRetryTimes)
		})
	}
}

type counterRoundTripperWare struct {
	count int
}

func (c *counterRoundTripperWare) Tripper() httptripperware.Tripperware {
	return func(next http.RoundTripper) http.RoundTripper {
		c.count = 0
		return httptripperware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			c.count++
			return next.RoundTrip(req)
		})
	}
}
