package metrics

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
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

		doneHandleRequest(InboundCall, ctx.Request.Method, ctx.FullPath(), httpStatus, elapsedTime)
	}
}
