package log

import (
	"context"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"time"

	"github.com/goccy/go-json"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type spanLogger struct {
	logger     *zap.Logger
	spanFields []zapcore.Field
	span       trace.Span
}

func (sl spanLogger) For(ctx context.Context, contextFields ...Fn) Logger {
	return sl
}

// With creates a child logger, and optionally adds some context fields to that logger.
func (sl spanLogger) With(fields ...zapcore.Field) Logger {
	return spanLogger{logger: sl.logger.With(fields...), span: sl.span, spanFields: sl.spanFields}
}

func (sl spanLogger) Sugar() *zap.SugaredLogger {
	return sl.logger.Sugar()
}

func (sl spanLogger) DPanic(msg string, fields ...zap.Field) {
	sl.logToSpan("panic", msg, fields...)
	sl.logger.DPanic(msg, append(sl.spanFields, fields...)...)
}

func (sl spanLogger) Debug(msg string, fields ...zapcore.Field) {
	sl.logToSpan("debug", msg, fields...)
	sl.logger.Debug(msg, append(sl.spanFields, fields...)...)
}

func (sl spanLogger) Info(msg string, fields ...zapcore.Field) {
	sl.logToSpan("info", msg, fields...)
	sl.logger.Info(msg, append(sl.spanFields, fields...)...)
}

func (sl spanLogger) Warn(msg string, fields ...zapcore.Field) {
	sl.logToSpan("warn", msg, fields...)
	sl.logger.Warn(msg, append(sl.spanFields, fields...)...)
}

func (sl spanLogger) Error(msg string, fields ...zapcore.Field) {
	sl.logToSpan("error", msg, fields...)
	sl.logger.Error(msg, append(sl.spanFields, fields...)...)
}

func (sl spanLogger) Fatal(msg string, fields ...zapcore.Field) {
	sl.logToSpan("fatal", msg, fields...)
	sl.logger.Fatal(msg, append(sl.spanFields, fields...)...)
}

func (sl spanLogger) Panic(msg string, fields ...zapcore.Field) {
	sl.logToSpan("panic", msg, fields...)
	sl.logger.Panic(msg, append(sl.spanFields, fields...)...)
}

func (sl spanLogger) logToSpan(level string, msg string, fields ...zapcore.Field) {
	attributes := []attribute.KeyValue{
		attribute.Key("level").String(level),
	}

	for _, field := range fields {
		attributes = appendField(attributes, field)
	}
	sl.span.AddEvent(msg, trace.WithAttributes(attributes...))
}

func appendField(attrs []attribute.KeyValue, f zapcore.Field) []attribute.KeyValue {
	switch f.Type {
	case zapcore.BoolType:
		attr := attribute.Bool(f.Key, f.Integer == 1)
		return append(attrs, attr)

	case zapcore.Int8Type, zapcore.Int16Type, zapcore.Int32Type, zapcore.Int64Type,
		zapcore.Uint32Type, zapcore.Uint8Type, zapcore.Uint16Type, zapcore.Uint64Type,
		zapcore.UintptrType, zapcore.DurationType, zapcore.TimeType:
		attr := attribute.Int64(f.Key, f.Integer)
		return append(attrs, attr)

	case zapcore.Float32Type, zapcore.Float64Type:
		var uintValue uint64
		if f.Integer >= 0 {
			uintValue = uint64(f.Integer)
		}
		attr := attribute.Float64(f.Key, math.Float64frombits(uintValue))
		return append(attrs, attr)

	case zapcore.Complex64Type:
		s := strconv.FormatComplex(complex128(f.Interface.(complex64)), 'E', -1, 64)
		attr := attribute.String(f.Key, s)
		return append(attrs, attr)
	case zapcore.Complex128Type:
		s := strconv.FormatComplex(f.Interface.(complex128), 'E', -1, 128)
		attr := attribute.String(f.Key, s)
		return append(attrs, attr)

	case zapcore.StringType:
		attr := attribute.String(f.Key, f.String)
		return append(attrs, attr)
	case zapcore.BinaryType, zapcore.ByteStringType:
		attr := attribute.String(f.Key, string(f.Interface.([]byte)))
		return append(attrs, attr)
	case zapcore.StringerType:
		attr := attribute.String(f.Key, f.Interface.(fmt.Stringer).String())
		return append(attrs, attr)

	case zapcore.TimeFullType:
		attr := attribute.Int64(f.Key, f.Interface.(time.Time).UnixNano())
		return append(attrs, attr)
	case zapcore.ErrorType:
		err := f.Interface.(error)
		typ := reflect.TypeOf(err).String()
		attrs = append(attrs, semconv.ExceptionTypeKey.String(typ))
		attrs = append(attrs, semconv.ExceptionMessageKey.String(err.Error()))
		return attrs
	case zapcore.ReflectType:
		attr := Attribute(f.Key, f.Interface)
		return append(attrs, attr)
	case zapcore.SkipType:
		return attrs

	case zapcore.ArrayMarshalerType:
		var attr attribute.KeyValue
		arrayEncoder := &bufferArrayEncoder{
			stringsSlice: []string{},
		}
		err := f.Interface.(zapcore.ArrayMarshaler).MarshalLogArray(arrayEncoder)
		if err != nil {
			attr = attribute.String(f.Key+"_error", fmt.Sprintf("otelzap: unable to marshal array: %v", err))
		} else {
			attr = attribute.StringSlice(f.Key, arrayEncoder.stringsSlice)
		}
		return append(attrs, attr)

	case zapcore.ObjectMarshalerType:
		attr := attribute.String(f.Key+"_error", "otelzap: zapcore.ObjectMarshalerType is not implemented")
		return append(attrs, attr)

	default:
		attr := attribute.String(f.Key+"_error", fmt.Sprintf("otelzap: unknown field type: %v", f))
		return append(attrs, attr)
	}
}

// bufferArrayEncoder implements zapcore.bufferArrayEncoder.
// It represents all added objects to their string values and
// adds them to the stringsSlice buffer.
type bufferArrayEncoder struct {
	stringsSlice []string
}

var _ zapcore.ArrayEncoder = (*bufferArrayEncoder)(nil)

func (t *bufferArrayEncoder) AppendComplex128(v complex128) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendComplex64(v complex64) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendArray(v zapcore.ArrayMarshaler) error {
	enc := &bufferArrayEncoder{}
	err := v.MarshalLogArray(enc)
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", enc.stringsSlice))
	return err
}

func (t *bufferArrayEncoder) AppendObject(v zapcore.ObjectMarshaler) error {
	m := zapcore.NewMapObjectEncoder()
	err := v.MarshalLogObject(m)
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", m.Fields))
	return err
}

func (t *bufferArrayEncoder) AppendReflected(v interface{}) error {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
	return nil
}

func (t *bufferArrayEncoder) AppendBool(v bool) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendByteString(v []byte) {
	t.stringsSlice = append(t.stringsSlice, string(v))
}

func (t *bufferArrayEncoder) AppendDuration(v time.Duration) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendFloat64(v float64) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendFloat32(v float32) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendInt(v int) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendInt64(v int64) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendInt32(v int32) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendInt16(v int16) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendInt8(v int8) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendString(v string) {
	t.stringsSlice = append(t.stringsSlice, v)
}

func (t *bufferArrayEncoder) AppendTime(v time.Time) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendUint(v uint) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendUint64(v uint64) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendUint32(v uint32) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendUint16(v uint16) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendUint8(v uint8) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func (t *bufferArrayEncoder) AppendUintptr(v uintptr) {
	t.stringsSlice = append(t.stringsSlice, fmt.Sprintf("%v", v))
}

func Attribute(key string, value interface{}) attribute.KeyValue {
	switch value := value.(type) {
	case nil:
		return attribute.String(key, "<nil>")
	case string:
		return attribute.String(key, value)
	case int:
		return attribute.Int(key, value)
	case int64:
		return attribute.Int64(key, value)
	case uint64:
		return attribute.Int64(key, int64(value))
	case float64:
		return attribute.Float64(key, value)
	case bool:
		return attribute.Bool(key, value)
	case fmt.Stringer:
		return attribute.String(key, value.String())
	}

	rv := reflect.ValueOf(value)

	switch rv.Kind() {
	case reflect.Array:
		rv = rv.Slice(0, rv.Len())
		fallthrough
	case reflect.Slice:
		switch reflect.TypeOf(value).Elem().Kind() {
		case reflect.Bool:
			return attribute.BoolSlice(key, rv.Interface().([]bool))
		case reflect.Int:
			return attribute.IntSlice(key, rv.Interface().([]int))
		case reflect.Int64:
			return attribute.Int64Slice(key, rv.Interface().([]int64))
		case reflect.Float64:
			return attribute.Float64Slice(key, rv.Interface().([]float64))
		case reflect.String:
			return attribute.StringSlice(key, rv.Interface().([]string))
		default:
			return attribute.KeyValue{Key: attribute.Key(key)}
		}
	case reflect.Bool:
		return attribute.Bool(key, rv.Bool())
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return attribute.Int64(key, rv.Int())
	case reflect.Float64:
		return attribute.Float64(key, rv.Float())
	case reflect.String:
		return attribute.String(key, rv.String())
	}
	if b, err := json.Marshal(value); b != nil && err == nil {
		return attribute.String(key, string(b))
	}
	return attribute.String(key, fmt.Sprint(value))
}
