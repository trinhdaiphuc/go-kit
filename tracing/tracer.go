package tracing

import (
	"context"
	"runtime"
	"time"

	"go.opentelemetry.io/contrib/propagators/b3"
	jaegerprob "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
)

// TracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Otel exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
//
// By default, if an environment variable is not set, and this option is not
// passed, "localhost:4317" will be used.
//
// If the OTEL_EXPORTER_OTLP_ENDPOINT or OTEL_EXPORTER_OTLP_METRICS_ENDPOINT
// environment variable is set, and this option is not passed, that variable
// value will be used. If both are set, OTEL_EXPORTER_OTLP_TRACES_ENDPOINT
// will take precedence.
func TracerProvider(serviceName, version string, opts ...otlptracegrpc.Option) (*tracesdk.TracerProvider, func(), error) {
	// Create the Otel exporter
	options := []otlptracegrpc.Option{
		otlptracegrpc.WithInsecure(),
	}
	traceClient := otlptracegrpc.NewClient(append(options, opts...)...)
	exp, err := otlptrace.New(context.Background(), traceClient)
	if err != nil {
		return nil, nil, err
	}

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exp),
		// Record information about this application in a Resource.
		tracesdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
				semconv.ServiceVersionKey.String(version),
			),
		),
	)

	// Register our TracerProvider as the global so any imported
	// instrumentation in the future will default to using it.
	otel.SetTracerProvider(tp)
	pb := propagation.NewCompositeTextMapPropagator(
		b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader|b3.B3SingleHeader)),
		propagation.TraceContext{},
		propagation.Baggage{},
		jaegerprob.Jaeger{},
	)
	otel.SetTextMapPropagator(pb)

	// Cleanly shutdown and flush telemetry when the application exits.
	cleanup := func() {
		// Do not make the application hang when it is shutdown.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		if err := tp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}

	return tp, cleanup, nil
}

func GetCaller(shift int) (funcName, file string, line int) {
	pc, file, line, ok := runtime.Caller(shift)
	if !ok {
		return
	}
	funcName = runtime.FuncForPC(pc).Name()
	return
}

// CreateSpan for each tracing layer
func CreateSpan(ctx context.Context, spanName string, spanOptions ...trace.SpanStartOption) (context.Context, trace.Span) {
	funcName, file, line := GetCaller(2)
	tracer := otel.Tracer(spanName)
	spanOptions = append(
		spanOptions,
		trace.WithAttributes(
			attribute.String("functionName", funcName),
			attribute.String("file", file),
			attribute.Int("line", line),
		),
	)
	ctx, span := tracer.Start(ctx, spanName, spanOptions...)
	return ctx, span
}

func GetTraceID(ctx context.Context) string {
	if span := trace.SpanFromContext(ctx); span != nil {
		spanCtx := span.SpanContext()
		return spanCtx.TraceID().String()
	}
	return ""
}
