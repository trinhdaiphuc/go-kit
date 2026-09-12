package tracing

import (
	"context"
	"encoding/json"
	"testing"

	"go.opentelemetry.io/otel/trace"
)

// The header chat-server puts on a Redis message must rebuild the same trace on
// the Go side; both ends run a composite b3 + traceparent propagator.
func TestExtractHeaderFromNodeProducer(t *testing.T) {
	if _, _, err := TracerProvider("test", "0", &OtelExporter{OTLPEndpoint: "127.0.0.1:1"}); err != nil {
		t.Fatalf("tracer provider: %v", err)
	}

	const nodeHeader = `{
		"traceparent":"00-d899d01152320464edb8dace9ed91a42-68b35a0eb8f93aed-01",
		"b3":"d899d01152320464edb8dace9ed91a42-68b35a0eb8f93aed-1",
		"x-b3-traceid":"d899d01152320464edb8dace9ed91a42",
		"x-b3-spanid":"68b35a0eb8f93aed",
		"x-b3-sampled":"1"
	}`

	header := map[string]string{}
	if err := json.Unmarshal([]byte(nodeHeader), &header); err != nil {
		t.Fatalf("unmarshal header: %v", err)
	}

	sc := trace.SpanContextFromContext(ExtractHeader(context.Background(), header))
	if got := sc.TraceID().String(); got != "d899d01152320464edb8dace9ed91a42" {
		t.Fatalf("trace id = %q", got)
	}
	if got := sc.SpanID().String(); got != "68b35a0eb8f93aed" {
		t.Fatalf("span id = %q", got)
	}
	if !sc.IsRemote() || !sc.IsSampled() {
		t.Fatalf("span context remote=%v sampled=%v", sc.IsRemote(), sc.IsSampled())
	}

	// Round trip: what Go injects must be readable by the same propagators.
	back := InjectHeader(trace.ContextWithRemoteSpanContext(context.Background(), sc))
	if back["traceparent"] == "" || back["b3"] == "" {
		t.Fatalf("inject produced %v", back)
	}
}
