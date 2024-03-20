package tracing

import (
	"context"
	"fmt"
	"path/filepath"
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

const (
	Resolver   = "Resolver"
	Service    = "Service"
	Repository = "Repository"
)

type Jaeger struct {
	AgentHost string `json:"agent_host,omitempty" mapstructure:"agent_host"`
	AgentPort string `json:"agent_port,omitempty" mapstructure:"agent_port"`
}

type OtelExporter struct {
	Jaeger *Jaeger `json:"jaeger,omitempty" mapstructure:"jaeger"`
}

// TracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Otel exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func TracerProvider(serviceName, version string, cfg *OtelExporter) (*tracesdk.TracerProvider, func(), error) {
	// Create the Otel exporter
	traceClient := otlptracegrpc.NewClient(
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithEndpoint(cfg.Jaeger.AgentHost+":"+cfg.Jaeger.AgentPort),
	)
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

	ctx, cancel := context.WithCancel(context.Background())

	cleanup := func(ctx context.Context) func() {
		return func() {
			defer cancel()

			// Cleanly shutdown and flush telemetry when the application exits.
			func(ctx context.Context) {
				// Do not make the application hang when it is shutdown.
				ctx, cancel = context.WithTimeout(ctx, time.Second*5)
				defer cancel()
				if err := tp.Shutdown(ctx); err != nil {
					otel.Handle(err)
				}
			}(ctx)
		}
	}(ctx)

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
func CreateSpan(ctx context.Context, spanName string) (context.Context, trace.Span) {
	funcName, file, line := GetCaller(2)
	tracer := otel.Tracer(fmt.Sprintf("%s:%d", file, line))
	ctx, span := tracer.Start(
		ctx,
		fmt.Sprintf("%s.%s", spanName, filepath.Base(funcName)),
		trace.WithAttributes(
			attribute.String("functionName", funcName),
			attribute.String("file", file),
			attribute.Int("line", line),
		),
		trace.WithSpanKind(trace.SpanKindServer),
	)
	return ctx, span
}

func GetTraceID(ctx context.Context) string {
	if span := trace.SpanFromContext(ctx); span != nil {
		spanCtx := span.SpanContext()
		return spanCtx.TraceID().String()
	}
	return ""
}
