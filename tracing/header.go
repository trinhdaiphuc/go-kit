package tracing

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
)

type headerCtxKey struct{}

// ContextWithHeader stores the header of the message being processed in ctx.
// Transports that strip the header before handing the message to the handler
// (Redis stream, ...) use it to hand the trace metadata over.
func ContextWithHeader(ctx context.Context, header map[string]string) context.Context {
	return context.WithValue(ctx, headerCtxKey{}, header)
}

// HeaderFromContext returns the header stored by ContextWithHeader, nil when
// the producer sent none.
func HeaderFromContext(ctx context.Context) map[string]string {
	header, _ := ctx.Value(headerCtxKey{}).(map[string]string)
	return header
}

// InjectHeader serialises the span context in ctx into a flat string map, ready
// to be carried as a message header (Redis stream field, pub/sub envelope, ...).
// It returns nil when ctx holds no span context worth propagating.
func InjectHeader(ctx context.Context) map[string]string {
	carrier := propagation.MapCarrier{}
	otel.GetTextMapPropagator().Inject(ctx, carrier)
	if len(carrier) == 0 {
		return nil
	}

	return carrier
}

// ExtractHeader rebuilds the remote span context from a message header produced
// by InjectHeader (or by any W3C/B3 compatible producer).
func ExtractHeader(ctx context.Context, header map[string]string) context.Context {
	if len(header) == 0 {
		return ctx
	}

	return otel.GetTextMapPropagator().Extract(ctx, propagation.MapCarrier(header))
}
