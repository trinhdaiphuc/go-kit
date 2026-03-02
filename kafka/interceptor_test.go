package kafka

import (
	"context"
	"testing"

	"github.com/IBM/sarama"
	"github.com/stretchr/testify/assert"
)

func TestLoggingInterceptor_OnSend(t *testing.T) {
	interceptor := NewLoggingInterceptor()

	msg := &sarama.ProducerMessage{
		Topic:     "test-topic",
		Key:       sarama.StringEncoder("key"),
		Value:     sarama.StringEncoder("value"),
		Partition: 1,
		Offset:    100,
	}

	// Should not panic
	interceptor.OnSend(msg)
}

func TestLoggingInterceptor_OnConsume(t *testing.T) {
	interceptor := NewLoggingInterceptor()

	msg := &sarama.ConsumerMessage{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    1,
		Key:       []byte("key"),
		Value:     []byte("value"),
	}

	// Should not panic - success case
	interceptor.OnConsume(msg)

	// Should not panic - error case
	interceptor.OnConsume(msg)
}

func TestProducerMessageCarrier(t *testing.T) {
	msg := &sarama.ProducerMessage{
		Topic: "test-topic",
		Headers: []sarama.RecordHeader{
			{Key: []byte("existing-key"), Value: []byte("existing-value")},
		},
	}

	carrier := NewProducerMessageCarrier(msg)

	// Test Get
	assert.Equal(t, "existing-value", carrier.Get("existing-key"))
	assert.Equal(t, "", carrier.Get("non-existent-key"))

	// Test Set
	carrier.Set("new-key", "new-value")
	assert.Equal(t, "new-value", carrier.Get("new-key"))

	// Test Set overwrites existing key
	carrier.Set("existing-key", "updated-value")
	assert.Equal(t, "updated-value", carrier.Get("existing-key"))

	// Test Keys
	keys := carrier.Keys()
	assert.Contains(t, keys, "new-key")
	assert.Contains(t, keys, "existing-key")
}

func TestConsumerMessageCarrier(t *testing.T) {
	msg := &sarama.ConsumerMessage{
		Topic: "test-topic",
		Headers: []*sarama.RecordHeader{
			{Key: []byte("existing-key"), Value: []byte("existing-value")},
		},
	}

	carrier := NewConsumerMessageCarrier(msg)

	// Test Get
	assert.Equal(t, "existing-value", carrier.Get("existing-key"))
	assert.Equal(t, "", carrier.Get("non-existent-key"))

	// Test Set
	carrier.Set("new-key", "new-value")
	assert.Equal(t, "new-value", carrier.Get("new-key"))

	// Test Set overwrites existing key
	carrier.Set("existing-key", "updated-value")
	assert.Equal(t, "updated-value", carrier.Get("existing-key"))

	// Test Keys
	keys := carrier.Keys()
	assert.Contains(t, keys, "new-key")
	assert.Contains(t, keys, "existing-key")
}

func TestConsumerMessageCarrier_NilHeaders(t *testing.T) {
	msg := &sarama.ConsumerMessage{
		Topic: "test-topic",
		Headers: []*sarama.RecordHeader{
			nil,
			{Key: []byte("key"), Value: []byte("value")},
		},
	}

	carrier := NewConsumerMessageCarrier(msg)

	// Should handle nil headers gracefully
	assert.Equal(t, "value", carrier.Get("key"))
	assert.Equal(t, "", carrier.Get("nil-key"))
}

func TestConsumerInterceptor_ChainManually(t *testing.T) {
	called := []string{}

	handler := func(ctx context.Context, msg *sarama.ConsumerMessage) error {
		called = append(called, "handler")
		return nil
	}

	interceptor1 := func(next ConsumerHandlerFn) ConsumerHandlerFn {
		return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
			called = append(called, "i1")
			return next(ctx, msg)
		}
	}

	interceptor2 := func(next ConsumerHandlerFn) ConsumerHandlerFn {
		return func(ctx context.Context, msg *sarama.ConsumerMessage) error {
			called = append(called, "i2")
			return next(ctx, msg)
		}
	}

	// Manual chaining: interceptor2(interceptor1(handler))
	// Execution order: i2 -> i1 -> handler
	wrapped := interceptor2(interceptor1(handler))
	err := wrapped(context.Background(), &sarama.ConsumerMessage{})

	assert.NoError(t, err)
	assert.Equal(t, []string{"i2", "i1", "handler"}, called)
}
