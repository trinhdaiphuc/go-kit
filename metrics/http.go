package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	httptripperware "github.com/trinhdaiphuc/go-kit/http/tripperware"
)

func MetricMiddleware(opts ...Option) gin.HandlerFunc {
	cfg := config{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	return func(ctx *gin.Context) {
		for _, f := range cfg.Filters {
			if !f(ctx.Request) {
				// Serve the request to the next middleware
				// if a filter rejects the request.
				ctx.Next()
				return
			}
		}
		startTime := time.Now()
		ctx.Next()
		elapsedTime := time.Since(startTime).Seconds()

		httpStatus := strconv.Itoa(ctx.Writer.Status())

		doneHTTPHandleRequest(InboundCall, ctx.Request.Method, ctx.FullPath(), httpStatus, elapsedTime)
	}
}

func ClientHTTPTripperware(opts ...Option) httptripperware.Tripperware {
	cfg := config{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	return func(next http.RoundTripper) http.RoundTripper {
		return httptripperware.RoundTripperFunc(func(req *http.Request) (*http.Response, error) {
			for _, f := range cfg.Filters {
				if !f(req) {
					return next.RoundTrip(req)
				}
			}

			startTime := time.Now()
			statusCode := http.StatusOK
			resp, err := next.RoundTrip(req)
			elapsedTime := time.Since(startTime).Seconds()

			if err != nil {
				statusCode = http.StatusInternalServerError
			}
			if resp != nil {
				statusCode = resp.StatusCode
			}

			httpStatus := strconv.Itoa(statusCode)

			endpoint := httpEndpoint(req.URL.Path, cfg.ServiceName)

			doneHTTPHandleRequest(OutboundCall, req.Method, endpoint, httpStatus, elapsedTime)
			return resp, err
		})
	}
}

func httpEndpoint(reqPath, serviceName string) string {
	return reqPath + " (" + serviceName + ")"
}
