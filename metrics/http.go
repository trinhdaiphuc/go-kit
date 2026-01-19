package metrics

import (
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/goccy/go-json"
	"github.com/grpc-ecosystem/grpc-gateway/v2/runtime"
	"google.golang.org/grpc/codes"

	"github.com/trinhdaiphuc/go-kit/errorx"
	"github.com/trinhdaiphuc/go-kit/http/middleware"
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

		doneHandleRequest(InboundCall, ctx.Request.Method, ctx.FullPath(), httpStatus, httpStatus, elapsedTime)
	}
}

func MetricHTTPMiddleware(opts ...Option) middleware.Middleware {
	cfg := config{}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			for _, f := range cfg.Filters {
				if !f(r) {
					next.ServeHTTP(w, r)
					return
				}
			}

			startTime := time.Now()
			resp := &responseWrapper{
				ResponseWriter: w,
				status:         http.StatusOK,
			}
			next.ServeHTTP(resp, r)
			elapsedTime := time.Since(startTime).Seconds()

			httpStatus := strconv.Itoa(resp.status)
			status := httpStatus

			// Read status code and body from response
			if resp.status == http.StatusOK && len(resp.body) > 0 {
				errResp := &errorResponse{}
				err := json.Unmarshal(resp.body, errResp)
				if err == nil && errResp != nil {
					if errResp.Error.Code <= http.StatusOK { // GRPC code
						code := codes.Code(errResp.Error.Code)
						status = code.String()
						httpStatus = strconv.Itoa(runtime.HTTPStatusFromCode(code))
					} else {
						httpStatus = strconv.Itoa(errResp.Error.Code)
						status = strconv.Itoa(errResp.Error.Code)
					}
				}
			}

			doneHandleRequest(InboundCall, r.Method, r.URL.Path, status, httpStatus, elapsedTime)
		})
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

			pattern := req.Pattern
			if pattern == "" {
				pattern = req.URL.Path
			}

			endpoint := httpEndpoint(pattern, cfg.ServiceName)

			doneHandleRequest(OutboundCall, req.Method, endpoint, httpStatus, httpStatus, elapsedTime)
			return resp, err
		})
	}
}

func httpEndpoint(reqPath, serviceName string) string {
	return reqPath + " (" + serviceName + ")"
}

type responseWrapper struct {
	http.ResponseWriter
	body   []byte
	status int
}

func (rw *responseWrapper) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func (rw *responseWrapper) Write(b []byte) (int, error) {
	rw.body = b
	return rw.ResponseWriter.Write(b)
}

func (rw *responseWrapper) Header() http.Header {
	return rw.ResponseWriter.Header()
}

type errorResponse struct {
	Error errorx.ErrorBody `json:"error,omitempty"`
}
