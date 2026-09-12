package redisstream

import (
	"context"
	"fmt"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/golang-queue/queue/core"
	"github.com/golang-queue/queue/job"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	"go.opentelemetry.io/contrib/propagators/b3"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"

	"github.com/trinhdaiphuc/go-kit/tracing"
)

// setupTracing installs a recording provider plus the same propagator the real
// TracerProvider() uses, so the test exercises the production wire format.
func setupTracing(t *testing.T) *tracetest.SpanRecorder {
	recorder := tracetest.NewSpanRecorder()
	tp := sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(recorder))
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		b3.New(b3.WithInjectEncoding(b3.B3MultipleHeader|b3.B3SingleHeader)),
		propagation.TraceContext{},
	))
	t.Cleanup(func() {
		require.NoError(t, tp.Shutdown(context.Background()))
		otel.SetTracerProvider(noop.NewTracerProvider())
	})

	return recorder
}

// Go service -> redis stream -> Go service: the consumer must end up in the
// trace the producer published with.
func TestTracePropagatesGoToGo(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	run := func(t *testing.T, stream string, producerCtx func() (context.Context, trace.Span)) (sent, received trace.TraceID) {
		var producerSpanID trace.SpanID
		recorder := setupTracing(t)

		got := make(chan trace.SpanContext, 1)
		consumer := NewWorker(
			WithAddr(endpoint),
			WithStreamName(stream),
			WithBlockTime(100*time.Millisecond),
			WithRunFunc(tracing.WrapRedisStreamConsumer("g1", stream, func(ctx context.Context, _ core.TaskMessage) error {
				got <- trace.SpanContextFromContext(ctx)
				return nil
			})),
		)
		defer func() { assert.NoError(t, consumer.Shutdown()) }()
		consumer.startConsumer()

		producer := NewWorker(WithAddr(endpoint), WithStreamName(stream))
		defer func() { assert.NoError(t, producer.Shutdown()) }()

		pctx, span := producerCtx()
		require.NoError(t, producer.QueueWithContext(pctx, &job.Message{Body: []byte("hi"), Timeout: time.Minute}))
		if span != nil {
			span.End()
		}

		task, err := consumer.Request()
		require.NoError(t, err)
		require.NoError(t, consumer.Run(context.Background(), task))

		select {
		case sc := <-got:
			assert.True(t, sc.IsValid(), "consumer must run inside a trace")
			received = sc.TraceID()
		case <-time.After(time.Second):
			t.Fatal("run func was not called")
		}

		for _, s := range recorder.Ended() {
			if s.Name() == producerSpanName {
				sent = s.SpanContext().TraceID()
				producerSpanID = s.SpanContext().SpanID()
			}
		}

		// The header must make the consumer a child of the producer span, not a
		// new root.
		for _, s := range recorder.Ended() {
			if s.Name() == "redis-stream-consumer" {
				assert.Equal(t, producerSpanID, s.Parent().SpanID(), "consumer must hang off the producer span")
				assert.True(t, s.Parent().IsRemote(), "parent must come from the message header")
			}
		}

		return sent, received
	}

	t.Run("no trace in context, the stream creates one", func(t *testing.T) {
		sent, received := run(t, "trace-fresh", func() (context.Context, trace.Span) {
			return context.Background(), nil
		})
		assert.True(t, sent.IsValid())
		assert.Equal(t, sent, received)
	})

	t.Run("caller already has a trace", func(t *testing.T) {
		var caller trace.TraceID
		sent, received := run(t, "trace-existing", func() (context.Context, trace.Span) {
			ctx, span := otel.Tracer("test").Start(context.Background(), "caller")
			caller = span.SpanContext().TraceID()
			return ctx, span
		})
		assert.Equal(t, caller, sent, "producer span must join the caller trace")
		assert.Equal(t, caller, received, "consumer must join the caller trace")
	})
}

// The header map is keyed by the *job.Message pointer; concurrent Request/Run
// pairs must never hand a message someone else's header.
func TestHeaderMapUnderConcurrency(t *testing.T) {
	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	const n = 100
	type mismatch struct{ want, got string }
	var (
		mu   sync.Mutex
		bad  []mismatch
		seen = make(map[string]int, n)
		wg   sync.WaitGroup
	)

	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("concurrent-header"),
		WithBlockTime(100*time.Millisecond),
		WithRunFunc(func(ctx context.Context, task core.TaskMessage) error {
			want := string(task.Payload())
			got := tracing.HeaderFromContext(ctx)["x-id"]

			mu.Lock()
			defer mu.Unlock()
			seen[want]++
			if want != got {
				bad = append(bad, mismatch{want: want, got: got})
			}

			return nil
		}),
	)
	defer func() { assert.NoError(t, w.Shutdown()) }()
	w.startConsumer()

	for i := 0; i < n; i++ {
		id := strconv.Itoa(i)
		body := job.Message{Body: []byte(id), Timeout: time.Minute}
		require.NoError(t, w.rdb.XAdd(ctx, &redis.XAddArgs{
			Stream: "concurrent-header",
			Values: map[string]any{
				fieldBody:   string(body.Bytes()),
				fieldHeader: fmt.Sprintf(`{"x-id":%q}`, id),
			},
		}).Err())
	}

	for i := 0; i < n; i++ {
		task, err := w.Request()
		require.NoError(t, err)

		wg.Add(1)
		go func() {
			defer wg.Done()
			assert.NoError(t, w.Run(context.Background(), task))
		}()
	}
	wg.Wait()

	assert.Empty(t, bad, "header mapped to the wrong message")
	assert.Len(t, seen, n, "every message must run exactly once")

	// Nothing may be left behind, otherwise the map grows for the process lifetime.
	empty := true
	w.headers.Range(func(any, any) bool { empty = false; return false })
	assert.True(t, empty, "headers map must be drained")
}

func BenchmarkHeaderMapHandoff(b *testing.B) {
	w := &Worker{}
	header := map[string]string{"traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01"}

	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			task := &job.Message{Body: []byte("x")}
			w.headers.Store(task, header)
			if _, ok := w.headers.LoadAndDelete(task); !ok {
				b.Fatal("header lost")
			}
		}
	})
}

func BenchmarkParseHeader(b *testing.B) {
	const raw = `{"traceparent":"00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01","b3":"4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-1"}`
	for i := 0; i < b.N; i++ {
		if len(parseHeader(raw)) != 2 {
			b.Fatal("bad parse")
		}
	}
}

// Fixture produced by chat-server's buildNotificationTask(): the Node producer
// and this consumer must keep agreeing on the field names and the job.Message
// shape, so the entry is replayed verbatim.
func TestConsumesChatServerEntry(t *testing.T) {
	const (
		nodeBody   = `{"body":"eyJmcm9tVXNlciI6eyJpZCI6InUxIn0sInRvVXNlcnMiOlsidTIiXX0=","timeout":30000000000}`
		nodeHeader = `{"traceparent":"00-d899d01152320464edb8dace9ed91a42-68b35a0eb8f93aed-01","b3":"d899d01152320464edb8dace9ed91a42-68b35a0eb8f93aed-1"}`
	)

	setupTracing(t)

	ctx := context.Background()
	redisC, endpoint := setupRedisContainer(ctx, t)
	defer testcontainers.CleanupContainer(t, redisC)

	type result struct {
		payload []byte
		traceID trace.TraceID
	}
	got := make(chan result, 1)
	w := NewWorker(
		WithAddr(endpoint),
		WithStreamName("from-node"),
		WithBlockTime(100*time.Millisecond),
		WithRunFunc(tracing.WrapRedisStreamConsumer("g1", "from-node", func(ctx context.Context, task core.TaskMessage) error {
			got <- result{payload: task.Payload(), traceID: trace.SpanContextFromContext(ctx).TraceID()}
			return nil
		})),
	)
	defer func() { assert.NoError(t, w.Shutdown()) }()
	w.startConsumer()

	require.NoError(t, w.rdb.XAdd(ctx, &redis.XAddArgs{
		Stream: "from-node",
		Values: map[string]any{fieldBody: nodeBody, fieldHeader: nodeHeader},
	}).Err())

	task, err := w.Request()
	require.NoError(t, err)
	msg, ok := task.(*job.Message)
	require.True(t, ok, "queue.run() only accepts *job.Message")
	assert.Equal(t, 30*time.Second, msg.Timeout)
	require.NoError(t, w.Run(context.Background(), task))

	select {
	case r := <-got:
		assert.JSONEq(t, `{"fromUser":{"id":"u1"},"toUsers":["u2"]}`, string(r.payload))
		assert.Equal(t, "d899d01152320464edb8dace9ed91a42", r.traceID.String())
	case <-time.After(time.Second):
		t.Fatal("run func was not called")
	}
}
