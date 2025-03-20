package tracing

import (
	"context"

	"github.com/golang-queue/queue/core"
	"go.opentelemetry.io/otel/attribute"
	semconv "go.opentelemetry.io/otel/semconv/v1.18.0"
	"go.opentelemetry.io/otel/trace"
)

type RedisStreamConsumer func(ctx context.Context, message core.TaskMessage) error

const (
	MessagingRedisConsumerGroupKey = attribute.Key("messaging.kafka.consumer.group")
)

func WrapRedisStreamConsumer(group, topic string, handler RedisStreamConsumer) RedisStreamConsumer {
	return func(ctx context.Context, message core.TaskMessage) error {
		ctx, span := CreateSpan(
			ctx, "redis-stream-consumer",
			trace.WithAttributes(
				semconv.MessagingSystem("redis-stream"),
				semconv.MessagingOperationReceive,
				semconv.MessagingDestinationKindTopic,
				MessagingRedisConsumerGroupKey.String(group),
				semconv.MessagingDestinationName(topic),
				semconv.MessagingMessagePayloadSizeBytes(len(message.Bytes())),
			),
			trace.WithSpanKind(trace.SpanKindConsumer),
		)
		defer span.End()

		err := handler(ctx, message)
		if err != nil {
			span.RecordError(err)
		}

		return err
	}
}
