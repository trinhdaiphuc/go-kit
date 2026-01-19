package middleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/log"
)

func WarningHighLatencyThreshold(highLatencyThreshold time.Duration) gin.HandlerFunc {
	return func(context *gin.Context) {
		start := time.Now()
		context.Next()
		elapsed := time.Since(start)
		if elapsed > highLatencyThreshold {
			log.For(context).Warn("[high-latency-threshold] this took too long", zap.Duration("latency", elapsed), zap.Duration("threshold", highLatencyThreshold))
		}
	}
}
