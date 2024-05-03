package tracing

import (
	"context"
	"fmt"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"

	"github.com/trinhdaiphuc/go-kit/kafka"
)

const (
	tracerName         = "zalopay.gitlab.com/bank-mapping-public-api-v2/pkg/tracing"
	kafkaMessagingName = "kafka"
)

// CreateSpanKafkaConsumerCtx create span for tracing message
func CreateSpanKafkaConsumerCtx(ctx context.Context, msg *sarama.ConsumerMessage, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	tracerProvider := otel.GetTracerProvider()
	tracer := tracerProvider.Tracer(
		tracerName,
		trace.WithInstrumentationVersion(sarama.DefaultVersion.String()),
	)

	opts = append(
		opts,
		trace.WithAttributes(
			semconv.MessagingSystem(kafkaMessagingName),
			semconv.MessagingSourceKindTopic,
			semconv.MessagingSourceName(msg.Topic),
			semconv.MessagingOperationReceive,
			semconv.MessagingKafkaSourcePartition(int(msg.Partition)),
			semconv.MessagingKafkaMessageOffset(int(msg.Offset)),
			semconv.MessagingKafkaMessageKey(string(msg.Key)),
			semconv.MessagingMessagePayloadSizeBytes(len(msg.Value)),
		),
		trace.WithSpanKind(trace.SpanKindConsumer),
	)

	return tracer.Start(ctx, fmt.Sprintf("%s receive", msg.Topic), opts...)
}

func WrapConsumerHandler(handler kafka.ConsumerHandlerFn) kafka.ConsumerHandlerFn {
	return func(ctx context.Context, message *sarama.ConsumerMessage) error {
		carrier := NewConsumerMessageCarrier(message)
		ctx = otel.GetTextMapPropagator().Extract(ctx, carrier)
		newCtx, span := CreateSpanKafkaConsumerCtx(ctx, message)
		defer span.End()

		err := handler(newCtx, message)
		if err != nil {
			span.SetAttributes(attribute.String("messaging.error", err.Error()))
		}

		return err
	}
}
