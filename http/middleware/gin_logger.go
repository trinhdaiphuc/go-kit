package middleware

import (
	"bytes"
	"io"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/trinhdaiphuc/go-kit/collection"
	"github.com/trinhdaiphuc/go-kit/log"
)

type Fn func(c *gin.Context) []zapcore.Field

// LoggerMiddlewareConfig is config setting for Ginzap
type LoggerMiddlewareConfig struct {
	SkipPaths []string
	Context   Fn
}

type bodyLogWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

const (
	maxBodyLogSize = 1 << 20 // 1 MB
)

var (
	bodyLogMethods = []string{http.MethodPost, http.MethodPut, http.MethodPatch}
)

// GinLoggerWithConfig returns a gin.HandlerFunc using configs
func GinLoggerWithConfig(conf *LoggerMiddlewareConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skipPaths[path] = true
	}

	return func(c *gin.Context) {
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		if _, ok := skipPaths[path]; ok {
			c.Next()
			return
		}

		var (
			fields = []zapcore.Field{
				zap.String("method", c.Request.Method),
				zap.String("query", c.Request.URL.RawQuery),
				zap.String("ip", c.ClientIP()),
				zap.String("user-agent", c.Request.UserAgent()),
			}
		)

		if collection.Contains(bodyLogMethods, c.Request.Method) {
			bodyByte, err := c.GetRawData()
			if err != nil {
				fields = append(fields, zap.Any("get_body_error", err))
			} else {
				c.Request.Body = io.NopCloser(bytes.NewReader(bodyByte))
				fields = append(fields, zap.ByteString("body", bodyByte))
			}
		}

		apiName := log.TrimHandler(c.HandlerName())
		c.Request = c.Request.WithContext(log.NewCtxLogger(c.Request.Context(), log.NewLogger(c.Request.Context(), zap.String("api_name", apiName))))

		// Wrapper body response for logging purposes
		blw := &bodyLogWriter{body: bytes.NewBufferString(""), ResponseWriter: c.Writer}
		c.Writer = blw

		start := time.Now()
		c.Next()
		latency := time.Since(start)

		fields = append(fields,
			zap.Int("status", c.Writer.Status()),
			zap.Duration("latency", latency),
			zap.Int("size", c.Writer.Size()),
		)

		if blw.body.Len() < maxBodyLogSize {
			fields = append(fields, zap.String("response", blw.body.String()))
		}

		if conf.Context != nil {
			fields = append(fields, conf.Context(c)...)
		}

		if len(c.Errors) > 0 {
			// Append error field if this is an erroneous request.
			fields = append(fields, zap.Reflect("error", c.Errors.JSON()))
			log.For(c).Error(path, fields...)
		} else {
			log.For(c).Info(path, fields...)
		}
	}
}

func (w *bodyLogWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}
