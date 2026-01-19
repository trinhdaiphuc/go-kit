package tracing

import (
	"context"
	"runtime"
	"time"

	"go.opentelemetry.io/contrib/propagators/b3"
	jaegerprob "go.opentelemetry.io/contrib/propagators/jaeger"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	tracesdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

const (
	SpanKindKey = attribute.Key("span.kind")
)

var (
	SchedulerSpanKind = SpanKindKey.String("scheduler")
)

type Jaeger struct {
	AgentHost string `json:"agent_host,omitempty" mapstructure:"agent_host"`
	AgentPort string `json:"agent_port,omitempty" mapstructure:"agent_port"`
	Endpoint  string `json:"endpoint,omitempty" mapstructure:"endpoint"`
	User      string `json:"user,omitempty" mapstructure:"user"`
	Password  string `json:"password,omitempty" mapstructure:"password"`
}

type OtelExporter struct {
	Jaeger       *Jaeger `json:"jaeger,omitempty" mapstructure:"jaeger"`
	OTLPEndpoint string  `json:"otlp_endpoint,omitempty" mapstructure:"otlp_endpoint"`
}

// TracerProvider returns an OpenTelemetry TracerProvider configured to use
// the Otel exporter that will send spans to the provided url. The returned
// TracerProvider will also use a Resource configured with all the information
// about the application.
func TracerProvider(serviceName, version string, cfg *OtelExporter) (trace.TracerProvider, func(), error) {
	var (
		exporter tracesdk.SpanExporter
		err      error
		ctx      = context.Background()
	)

	if cfg.Jaeger != nil {
		exporter, err = jaeger.New(
			// This will use the following environment variables for configuration if no explicit option is provided:
			jaeger.WithAgentEndpoint(
				jaeger.WithAgentHost(cfg.Jaeger.AgentHost),
				jaeger.WithAgentPort(cfg.Jaeger.AgentPort),
			),
		)
		if err != nil {
			return nil, nil, err
		}
	} else {
		// Create the Otel exporter
		traceClient := otlptracegrpc.NewClient(
			otlptracegrpc.WithInsecure(),
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
		)
		exporter, err = otlptrace.New(ctx, traceClient)
		if err != nil {
			return nil, nil, err
		}
	}

	res, err := resource.New(
		ctx,
		resource.WithFromEnv(),
		resource.WithProcess(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersion(version),
		),
	)
	if err != nil {
		return nil, nil, err
	}

	tp := tracesdk.NewTracerProvider(
		// Always be sure to batch in production.
		tracesdk.WithBatcher(exporter),
		// Record information about this application in a Resource.
		tracesdk.WithResource(res),
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
