package tracing

import (
	"context"
	"encoding/binary"
	"fmt"
	"strconv"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/trinhdaiphuc/go-kit/kafka"
)

type producer struct {
	kafka.Producer
	cfg          config
	saramaConfig *sarama.Config
}

// SendMessage calls sarama.SyncProducer.SendMessage and traces the request.
func (p *producer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	span := startProducerSpan(p.cfg, p.saramaConfig.Version, msg)
	partition, offset, err = p.Producer.SendMessage(msg)
	finishProducerSpan(span, partition, offset, err)
	return partition, offset, err
}

// SendMessages calls sarama.SyncProducer.SendMessages and traces the requests.
func (p *producer) SendMessages(msgs []*sarama.ProducerMessage) error {
	// Although there's only one call made to the SyncProducer, the messages are
	// treated individually, so we create a span for each one
	spans := make([]trace.Span, len(msgs))
	for i, msg := range msgs {
		spans[i] = startProducerSpan(p.cfg, p.saramaConfig.Version, msg)
	}
	err := p.Producer.SendMessages(msgs)
	for i, span := range spans {
		finishProducerSpan(span, msgs[i].Partition, msgs[i].Offset, err)
	}
	return err
}

// WrapKafkaProducer wraps a sarama.SyncProducer so that all produced messages
// are traced.
func WrapKafkaProducer(p kafka.Producer, opts ...Option) kafka.Producer {
	cfg := newConfig(opts...)

	return &producer{
		Producer:     p,
		cfg:          cfg,
		saramaConfig: p.GetClient().Config(),
	}
}

func startProducerSpan(cfg config, version sarama.KafkaVersion, msg *sarama.ProducerMessage) trace.Span {
	// If there's a span context in the message, use that as the parent context.
	carrier := NewProducerMessageCarrier(msg)
	ctx := cfg.Propagators.Extract(context.Background(), carrier)

	// Create a span.
	attrs := []attribute.KeyValue{
		semconv.MessagingSystem(kafkaMessagingName),
		semconv.MessagingDestinationKindTopic,
		semconv.MessagingDestinationName(msg.Topic),
		semconv.MessagingMessagePayloadSizeBytes(msgPayloadSize(msg, version)),
		semconv.MessagingOperationPublish,
		semconv.MessagingKafkaDestinationPartition(int(msg.Partition)),
		semconv.MessagingKafkaMessageOffset(int(msg.Offset)),
	}
	opts := []trace.SpanStartOption{
		trace.WithAttributes(attrs...),
		trace.WithSpanKind(trace.SpanKindProducer),
	}
	ctx, span := cfg.Tracer.Start(ctx, fmt.Sprintf("%s publish", msg.Topic), opts...)

	if version.IsAtLeast(sarama.V0_11_0_0) {
		// Inject current span context, so consumers can use it to propagate span.
		cfg.Propagators.Inject(ctx, carrier)
	}

	return span
}

func finishProducerSpan(span trace.Span, partition int32, offset int64, err error) {
	span.SetAttributes(
		semconv.MessagingMessageID(strconv.FormatInt(offset, 10)),
		semconv.MessagingKafkaDestinationPartition(int(partition)),
	)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
	}
	span.End()
}

type config struct {
	TracerProvider trace.TracerProvider
	Propagators    propagation.TextMapPropagator

	Tracer trace.Tracer
}

// newConfig returns a config with all Options set.
func newConfig(opts ...Option) config {
	cfg := config{
		Propagators:    otel.GetTextMapPropagator(),
		TracerProvider: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt.apply(&cfg)
	}

	cfg.Tracer = cfg.TracerProvider.Tracer(
		tracerName,
		trace.WithInstrumentationVersion(sarama.DefaultVersion.String()),
	)

	return cfg
}

// Option interface used for setting optional config properties.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (fn optionFunc) apply(c *config) {
	fn(c)
}

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(
		func(cfg *config) {
			if provider != nil {
				cfg.TracerProvider = provider
			}
		},
	)
}

// WithPropagators specifies propagators to use for extracting
// information from the HTTP requests. If none are specified, global
// ones will be used.
func WithPropagators(propagators propagation.TextMapPropagator) Option {
	return optionFunc(
		func(cfg *config) {
			if propagators != nil {
				cfg.Propagators = propagators
			}
		},
	)
}

// msgPayloadSize returns the approximate estimate of message size in bytes.
//
// For kafka version <= 0.10, the message size is
// ~ 4(crc32) + 8(timestamp) + 4(key len) + 4(value len) + 4(message len) + 1(attrs) + 1(magic).
//
// For kafka version >= 0.11, the message size with varint encoding is
// ~ 5 * (crc32, key len, value len, message len, attrs) + timestamp + 1 byte (magic).
// + header key + header value + header key len + header value len.
func msgPayloadSize(msg *sarama.ProducerMessage, kafkaVersion sarama.KafkaVersion) int {
	maximumRecordOverhead := 5*binary.MaxVarintLen32 + binary.MaxVarintLen64 + 1
	producerMessageOverhead := 26
	version := 1
	if kafkaVersion.IsAtLeast(sarama.V0_11_0_0) {
		version = 2
	}
	size := producerMessageOverhead
	if version >= 2 {
		size = maximumRecordOverhead
		for _, h := range msg.Headers {
			size += len(h.Key) + len(h.Value) + 2*binary.MaxVarintLen32
		}
	}
	if msg.Key != nil {
		size += msg.Key.Length()
	}
	if msg.Value != nil {
		size += msg.Value.Length()
	}
	return size
}
