package log

import (
	"context"
	"errors"
	"fmt"
	"math"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

func TestSpanLogger_Debug(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	span := trace.SpanFromContext(context.Background())
	sl := spanLogger{logger: logger, span: span}

	sl.Debug("debug message", zap.String("key", "value"))
}

func TestSpanLogger_Info(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	span := trace.SpanFromContext(context.Background())
	sl := spanLogger{logger: logger, span: span}

	sl.Info("info message", zap.String("key", "value"))
}

func TestSpanLogger_Warn(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	span := trace.SpanFromContext(context.Background())
	sl := spanLogger{logger: logger, span: span}

	sl.Warn("warn message", zap.String("key", "value"))
}

func TestSpanLogger_Error(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	span := trace.SpanFromContext(context.Background())
	sl := spanLogger{logger: logger, span: span}

	sl.Error("error message", zap.String("key", "value"))
}

type stringer struct{}

func (s stringer) String() string {
	return "stringer"
}

func TestAppendField(t *testing.T) {
	tests := []struct {
		name  string
		field zapcore.Field
		want  []attribute.KeyValue
	}{
		{
			name:  "BoolType",
			field: zapcore.Field{Key: "key", Type: zapcore.BoolType, Integer: 1},
			want:  []attribute.KeyValue{attribute.Bool("key", true)},
		},
		{
			name:  "Int64Type",
			field: zapcore.Field{Key: "key", Type: zapcore.Int64Type, Integer: 123},
			want:  []attribute.KeyValue{attribute.Int64("key", 123)},
		},
		{
			name:  "Float64Type",
			field: zapcore.Field{Key: "key", Type: zapcore.Float64Type, Integer: int64(math.Float64bits(1.23))},
			want:  []attribute.KeyValue{attribute.Float64("key", 1.23)},
		},
		{
			name:  "StringType",
			field: zapcore.Field{Key: "key", Type: zapcore.StringType, String: "value"},
			want:  []attribute.KeyValue{attribute.String("key", "value")},
		},
		{
			name:  "BinaryType",
			field: zapcore.Field{Key: "key", Type: zapcore.BinaryType, Interface: []byte("binary")},
			want:  []attribute.KeyValue{attribute.String("key", "binary")},
		},
		{
			name:  "ByteStringType",
			field: zapcore.Field{Key: "key", Type: zapcore.ByteStringType, Interface: []byte("bytestring")},
			want:  []attribute.KeyValue{attribute.String("key", "bytestring")},
		},
		{
			name: "StringerType",
			field: zapcore.Field{
				Key: "key", Type: zapcore.StringerType, Interface: stringer{},
			},
			want: []attribute.KeyValue{attribute.String("key", "stringer")},
		},
		{
			name:  "TimeFullType",
			field: zapcore.Field{Key: "key", Type: zapcore.TimeFullType, Interface: time.Unix(0, 0)},
			want:  []attribute.KeyValue{attribute.Int64("key", 0)},
		},
		{
			name:  "ErrorType",
			field: zapcore.Field{Key: "key", Type: zapcore.ErrorType, Interface: errors.New("error")},
			want: []attribute.KeyValue{
				semconv.ExceptionTypeKey.String("*errors.errorString"),
				semconv.ExceptionMessageKey.String("error"),
			},
		},
		{
			name:  "ReflectType",
			field: zapcore.Field{Key: "key", Type: zapcore.ReflectType, Interface: "reflect"},
			want:  []attribute.KeyValue{attribute.String("key", "reflect")},
		},
		{
			name:  "SkipType",
			field: zapcore.Field{Key: "key", Type: zapcore.SkipType},
			want:  nil,
		},
		{
			name: "ArrayMarshalerType",
			field: zapcore.Field{
				Key: "key", Type: zapcore.ArrayMarshalerType,
				Interface: zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
					enc.AppendString("array")
					return nil
				}),
			},
			want: []attribute.KeyValue{attribute.StringSlice("key", []string{"array"})},
		},
		{
			name: "ObjectMarshalerType",
			field: zapcore.Field{
				Key: "key", Type: zapcore.ObjectMarshalerType,
				Interface: zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
					enc.AddString("object", "value")
					return nil
				}),
			},
			want: []attribute.KeyValue{attribute.String("key_error", "otelzap: zapcore.ObjectMarshalerType is not implemented")},
		},
		{
			name:  "UnknownType",
			field: zapcore.Field{Key: "key", Type: zapcore.FieldType(33)},
			want:  []attribute.KeyValue{attribute.String("key_error", "otelzap: unknown field type: {key 33 0  <nil>}")},
		},
		{
			name:  "Complex64Type",
			field: zapcore.Field{Key: "key", Type: zapcore.Complex64Type, Interface: complex64(1 + 2i)},
			want:  []attribute.KeyValue{attribute.String("key", "(1E+00+2E+00i)")},
		},
		{
			name:  "Complex128Type",
			field: zapcore.Field{Key: "key", Type: zapcore.Complex128Type, Interface: complex128(3 + 4i)},
			want:  []attribute.KeyValue{attribute.String("key", "(3E+00+4E+00i)")},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := appendField(nil, tt.field)
			assert.Equalf(t, tt.want, got, "appendField() = %v, want %v", got, tt.want)
		})
	}
}

type Boolean bool
type Float64Type float64
type StringType string

func TestAttribute(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value interface{}
		want  attribute.KeyValue
	}{
		{
			name:  "Nil",
			key:   "key",
			value: nil,
			want:  attribute.String("key", "<nil>"),
		},
		{
			name:  "String",
			key:   "key",
			value: "value",
			want:  attribute.String("key", "value"),
		},
		{
			name:  "Int",
			key:   "key",
			value: 123,
			want:  attribute.Int("key", 123),
		},
		{
			name:  "Int64",
			key:   "key",
			value: int64(123),
			want:  attribute.Int64("key", 123),
		},
		{
			name:  "Uint64",
			key:   "key",
			value: uint64(123),
			want:  attribute.Int64("key", 123),
		},
		{
			name:  "Float64",
			key:   "key",
			value: 1.23,
			want:  attribute.Float64("key", 1.23),
		},
		{
			name:  "Bool",
			key:   "key",
			value: true,
			want:  attribute.Bool("key", true),
		},
		{
			name:  "Stringer",
			key:   "key",
			value: fmt.Stringer(stringer{}),
			want:  attribute.String("key", "stringer"),
		},
		{
			name:  "BoolSlice",
			key:   "key",
			value: []bool{true, false},
			want:  attribute.BoolSlice("key", []bool{true, false}),
		},
		{
			name:  "IntSlice",
			key:   "key",
			value: []int{1, 2, 3},
			want:  attribute.IntSlice("key", []int{1, 2, 3}),
		},
		{
			name:  "Int64Slice",
			key:   "key",
			value: []int64{1, 2, 3},
			want:  attribute.Int64Slice("key", []int64{1, 2, 3}),
		},
		{
			name:  "Float64Slice",
			key:   "key",
			value: []float64{1.1, 2.2, 3.3},
			want:  attribute.Float64Slice("key", []float64{1.1, 2.2, 3.3}),
		},
		{
			name:  "StringSlice",
			key:   "key",
			value: []string{"a", "b", "c"},
			want:  attribute.StringSlice("key", []string{"a", "b", "c"}),
		},
		{
			name:  "DefaultCase Slice",
			key:   "key",
			value: []struct{}{},
			want:  attribute.KeyValue{Key: attribute.Key("key")},
		},
		{
			name:  "Reflect",
			key:   "key",
			value: struct{ Name string }{"test"},
			want:  attribute.String("key", `{"Name":"test"}`),
		},
		{
			name:  "DefaultCase",
			key:   "key",
			value: struct{}{},
			want:  attribute.String("key", "{}"),
		},
		{
			name:  "Bool",
			key:   "key",
			value: true,
			want:  attribute.Bool("key", true),
		},
		{
			name:  "Bool (Boolean)",
			key:   "key",
			value: Boolean(true),
			want:  attribute.Bool("key", true),
		},
		{
			name:  "Int",
			key:   "key",
			value: 123,
			want:  attribute.Int64("key", 123),
		},
		{
			name:  "Int8",
			key:   "key",
			value: int8(123),
			want:  attribute.Int64("key", 123),
		},
		{
			name:  "Int16",
			key:   "key",
			value: int16(123),
			want:  attribute.Int64("key", 123),
		},
		{
			name:  "Int32",
			key:   "key",
			value: int32(123),
			want:  attribute.Int64("key", 123),
		},
		{
			name:  "Int64",
			key:   "key",
			value: int64(123),
			want:  attribute.Int64("key", 123),
		},
		{
			name:  "Float64",
			key:   "key",
			value: Float64Type(1.23),
			want:  attribute.Float64("key", 1.23),
		},
		{
			name:  "String",
			key:   "key",
			value: StringType("value"),
			want:  attribute.String("key", "value"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Attribute(tt.key, tt.value)
			assert.Equalf(t, tt.want, got, "Attribute() = %v, want %v", got, tt.want)
		})
	}
}

func TestBufferArrayEncoder(t *testing.T) {
	enc := &bufferArrayEncoder{}

	enc.AppendBool(true)
	assert.Equal(t, []string{"true"}, enc.stringsSlice)

	enc.AppendInt(123)
	assert.Equal(t, []string{"true", "123"}, enc.stringsSlice)

	enc.AppendString("test")
	assert.Equal(t, []string{"true", "123", "test"}, enc.stringsSlice)

	enc.AppendByteString([]byte("bytes"))
	assert.Equal(t, []string{"true", "123", "test", "bytes"}, enc.stringsSlice)

	enc.AppendDuration(time.Second)
	assert.Equal(t, []string{"true", "123", "test", "bytes", "1s"}, enc.stringsSlice)

	enc.AppendFloat64(1.23)
	assert.Equal(t, []string{"true", "123", "test", "bytes", "1s", "1.23"}, enc.stringsSlice)

	enc.AppendFloat32(4.56)
	assert.Equal(t, []string{"true", "123", "test", "bytes", "1s", "1.23", "4.56"}, enc.stringsSlice)

	enc.AppendInt64(789)
	assert.Equal(t, []string{"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789"}, enc.stringsSlice)

	enc.AppendInt32(456)
	assert.Equal(t, []string{"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456"}, enc.stringsSlice)

	enc.AppendInt16(1234)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234",
	}, enc.stringsSlice)

	enc.AppendInt8(12)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
	}, enc.stringsSlice)

	enc.AppendTime(time.Unix(0, 0).UTC())
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC",
	}, enc.stringsSlice)

	enc.AppendUint(123)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123",
	}, enc.stringsSlice)

	enc.AppendUint64(456)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456",
	}, enc.stringsSlice)

	enc.AppendUint32(789)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456", "789",
	}, enc.stringsSlice)

	enc.AppendUint16(1234)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456", "789", "1234",
	}, enc.stringsSlice)

	enc.AppendUint8(12)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456", "789", "1234", "12",
	}, enc.stringsSlice)

	enc.AppendUintptr(123)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456", "789", "1234", "12", "123",
	}, enc.stringsSlice)

	enc.AppendComplex128(1 + 2i)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456", "789", "1234", "12", "123", "(1+2i)",
	}, enc.stringsSlice)

	enc.AppendComplex64(3 + 4i)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456", "789", "1234", "12", "123", "(1+2i)", "(3+4i)",
	}, enc.stringsSlice)

	err := enc.AppendArray(zapcore.ArrayMarshalerFunc(func(enc zapcore.ArrayEncoder) error {
		enc.AppendString("array")
		return nil
	}))
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456", "789", "1234", "12", "123", "(1+2i)", "(3+4i)", "[array]",
	}, enc.stringsSlice)

	err = enc.AppendObject(zapcore.ObjectMarshalerFunc(func(enc zapcore.ObjectEncoder) error {
		enc.AddString("object", "value")
		return nil
	}))
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456", "789", "1234", "12", "123", "(1+2i)", "(3+4i)", "[array]",
		"map[object:value]",
	}, enc.stringsSlice)

	err = enc.AppendReflected("reflected")
	assert.NoError(t, err)
	assert.Equal(t, []string{
		"true", "123", "test", "bytes", "1s", "1.23", "4.56", "789", "456", "1234", "12",
		"1970-01-01 00:00:00 +0000 UTC", "123", "456", "789", "1234", "12", "123", "(1+2i)", "(3+4i)", "[array]",
		"map[object:value]", "reflected",
	}, enc.stringsSlice)
}
