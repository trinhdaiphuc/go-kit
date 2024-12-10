package log

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/k0kubun/pp"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// ConsoleEncoderName ...
const (
	ConsoleEncoderName = "custom_console"
)

var (
	instance *Factory
	once     = sync.Once{}
)

// Factory wraps zap.Logger
type Factory struct {
	*zap.Logger
	ll *zap.Logger
}

type Fn func(ctx context.Context) []zap.Field

type Config struct {
	Level   string `json:"level" yaml:"level" mapstructure:"level"`
	Encoder string `json:"encoder" yaml:"encoder" mapstructure:"encoder"`
}

type Logger interface {
	Sugar() *zap.SugaredLogger
	For(ctx context.Context, contextFields ...Fn) Logger
	With(fields ...zap.Field) Logger
	Debug(msg string, fields ...zap.Field)
	Info(msg string, fields ...zap.Field)
	Warn(msg string, fields ...zap.Field)
	Error(msg string, fields ...zap.Field)
	DPanic(msg string, fields ...zap.Field)
	Panic(msg string, fields ...zap.Field)
	Fatal(msg string, fields ...zap.Field)
}

func (logger Factory) Bg() Logger {
	return logger.With()
}

// PrintError prints all error with all metadata and line number.
// It's preferred to be used at top level function.
//
//	func DoSomething() (_err error) {
//	    defer instance.PrintError("DoSomething", &_err)
func (logger Factory) PrintError(msg string, err *error) {
	if *err != nil {
		instance.Sugar().Errorf("%v: %+v", msg, *err)
	}
}

// For context: add span and trace_id into logger fields
func (logger Factory) For(ctx context.Context, contextFields ...Fn) Logger {
	if span := trace.SpanFromContext(ctx); span != nil {
		spanCtx := span.SpanContext()
		fields := []zapcore.Field{
			zap.String("trace_id", spanCtx.TraceID().String()),
			zap.String("span_id", spanCtx.SpanID().String()),
		}

		for _, ctxField := range contextFields {
			fields = append(fields, ctxField(ctx)...)
		}

		return spanLogger{
			logger:     logger.ll,
			spanFields: fields,
			span:       span,
		}
	}
	return logger
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (logger Factory) With(fields ...zapcore.Field) Logger {
	return Factory{Logger: logger.Logger.With(fields...)}
}

// Shorthand functions for logging.
var (
	Any        = zap.Any
	Bool       = zap.Bool
	Duration   = zap.Duration
	Float64    = zap.Float64
	Int        = zap.Int
	Int64      = zap.Int64
	Skip       = zap.Skip
	String     = zap.String
	Strings    = zap.Strings
	Stringer   = zap.Stringer
	Time       = zap.Time
	Uint       = zap.Uint
	Uint32     = zap.Uint32
	Uint64     = zap.Uint64
	Uintptr    = zap.Uintptr
	ByteString = zap.ByteString
	Error      = zap.Error
	Reflect    = zap.Reflect
)

// DefaultConsoleEncoderConfig ...
var DefaultConsoleEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        "time",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "msg",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    zapcore.CapitalColorLevelEncoder,
	EncodeTime:     zapcore.ISO8601TimeEncoder,
	EncodeDuration: zapcore.StringDurationEncoder,
	EncodeCaller:   ShortColorCallerEncoder,
}

var StructureEncoderConfig = zapcore.EncoderConfig{
	TimeKey:        "timestamp",
	LevelKey:       "level",
	NameKey:        "logger",
	CallerKey:      "caller",
	MessageKey:     "message",
	StacktraceKey:  "stacktrace",
	LineEnding:     zapcore.DefaultLineEnding,
	EncodeLevel:    EncodeLevel,
	EncodeTime:     RFC3339NanoTimeEncoder,
	EncodeDuration: zapcore.SecondsDurationEncoder,
	EncodeCaller:   ShortColorCallerEncoder,
}

var logLevelSeverity = map[zapcore.Level]string{
	zapcore.DebugLevel:  "DEBUG",
	zapcore.InfoLevel:   "INFO",
	zapcore.WarnLevel:   "WARNING",
	zapcore.ErrorLevel:  "ERROR",
	zapcore.DPanicLevel: "CRITICAL",
	zapcore.PanicLevel:  "ALERT",
	zapcore.FatalLevel:  "EMERGENCY",
}

// EncodeLevel maps the internal Zap log level to the appropriate Stack driver
// level.
func EncodeLevel(l zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(logLevelSeverity[l])
}

// RFC3339NanoTimeEncoder serializes a time.Time to an RFC3339Nano-formatted
// string with nanoseconds precision.
func RFC3339NanoTimeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format(time.RFC3339Nano))
}

// Interface ...
func Interface(key string, val interface{}) zapcore.Field {
	if val, ok := val.(fmt.Stringer); ok {
		return zap.Stringer(key, val)
	}
	return zap.Reflect(key, val)
}

// Stack ...
func Stack() zapcore.Field {
	return zap.Stack("stack")
}

// Int32 ...
func Int32(key string, val int32) zapcore.Field {
	return zap.Int(key, int(val))
}

// Object ...
var Object = zap.Any

type dd struct {
	v interface{}
}

func (d dd) String() string {
	return pp.Sprint(d.v)
}

// Dump renders object for debugging
func Dump(v interface{}) fmt.Stringer {
	return dd{v}
}

func TrimHandler(handlerName string) string {
	result := filepath.Base(handlerName)
	names := strings.Split(result, ".")
	if len(names) > 0 {
		result = names[len(names)-1]
		result = strings.TrimSuffix(result, "-fm")
	}
	return result
}

// ShortColorCallerEncoder encodes caller information with sort path filename and enable color.
func ShortColorCallerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	callerStr := caller.TrimmedPath()
	enc.AppendString(callerStr)
}

func newLogger(cfg *Config, opts ...zap.Option) *zap.Logger {
	enabler, err := zap.ParseAtomicLevel(cfg.Level)
	if err != nil {
		panic(err)
	}

	loggerConfig := zap.Config{
		Level:       enabler,
		Development: false,
		Sampling: &zap.SamplingConfig{
			Initial:    100,
			Thereafter: 100,
		},
		Encoding:         "json",
		EncoderConfig:    StructureEncoderConfig,
		OutputPaths:      []string{"stderr"},
		ErrorOutputPaths: []string{"stderr"},
	}

	if cfg.Encoder == ConsoleEncoderName {
		loggerConfig = zap.Config{
			Level:            enabler,
			Development:      false,
			Encoding:         ConsoleEncoderName,
			EncoderConfig:    DefaultConsoleEncoderConfig,
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
		Object = func(key string, val interface{}) zap.Field {
			return zap.Stringer(key, Dump(val))
		}
	}
	stacktraceLevel := zap.NewAtomicLevelAt(zapcore.PanicLevel)

	opts = append(opts, zap.AddStacktrace(stacktraceLevel))
	logger, err := loggerConfig.Build(opts...)
	if err != nil {
		panic(err)
	}
	return logger
}

func New(cfg *Config) *Factory {
	once.Do(func() {
		err := zap.RegisterEncoder(ConsoleEncoderName, func(cfg zapcore.EncoderConfig) (zapcore.Encoder, error) {
			return NewConsoleEncoder(cfg), nil
		})
		if err != nil {
			panic(err)
		}
		instance = &Factory{
			Logger: newLogger(cfg, zap.AddCaller()),
			ll:     newLogger(cfg, zap.AddCallerSkip(1), zap.AddCaller()),
		}
	})

	return instance
}

// Bg creates a context-unaware logger.
func Bg() Logger {
	if instance == nil {
		New(&Config{
			Level:   "info",
			Encoder: "json",
		})
	}
	return instance
}

// For returns a context-aware Logger. If the context
// contains an OpenTracing span, all logging calls are also
// echo-ed into the span.
func For(ctx context.Context, contextFields ...Fn) Logger {
	if instance == nil {
		New(&Config{
			Level:   "info",
			Encoder: "json",
		})
	}
	return instance.For(ctx, contextFields...)
}
