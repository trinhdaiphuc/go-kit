package kafka

import (
	"context"

	"github.com/IBM/sarama"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/log"
)

// ConsumerInterceptor is a function that wraps a ConsumerHandlerFn with additional behavior.
type ConsumerInterceptor func(next ConsumerHandlerFn) ConsumerHandlerFn

// ProducerMessageCarrier injects and extracts traces from a sarama.ProducerMessage.
// Implements propagation.TextMapCarrier interface.
type ProducerMessageCarrier struct {
	msg *sarama.ProducerMessage
}

var _ propagation.TextMapCarrier = (*ProducerMessageCarrier)(nil)

// NewProducerMessageCarrier creates a new ProducerMessageCarrier.
func NewProducerMessageCarrier(msg *sarama.ProducerMessage) *ProducerMessageCarrier {
	return &ProducerMessageCarrier{msg: msg}
}

// Get retrieves a single value for a given key.
func (c *ProducerMessageCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set sets a header.
func (c *ProducerMessageCarrier) Set(key, val string) {
	// Ensure uniqueness of keys
	for i := 0; i < len(c.msg.Headers); i++ {
		if string(c.msg.Headers[i].Key) == key {
			c.msg.Headers = append(c.msg.Headers[:i], c.msg.Headers[i+1:]...)
			i--
		}
	}
	c.msg.Headers = append(
		c.msg.Headers, sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(val),
		},
	)
}

// Keys returns a slice of all key identifiers in the carrier.
func (c *ProducerMessageCarrier) Keys() []string {
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = string(h.Key)
	}
	return out
}

// ConsumerMessageCarrier injects and extracts traces from a sarama.ConsumerMessage.
// Implements propagation.TextMapCarrier interface.
type ConsumerMessageCarrier struct {
	msg *sarama.ConsumerMessage
}

var _ propagation.TextMapCarrier = (*ConsumerMessageCarrier)(nil)

// NewConsumerMessageCarrier creates a new ConsumerMessageCarrier.
func NewConsumerMessageCarrier(msg *sarama.ConsumerMessage) *ConsumerMessageCarrier {
	return &ConsumerMessageCarrier{msg: msg}
}

// Get retrieves a single value for a given key.
func (c *ConsumerMessageCarrier) Get(key string) string {
	for _, h := range c.msg.Headers {
		if h != nil && string(h.Key) == key {
			return string(h.Value)
		}
	}
	return ""
}

// Set sets a header.
func (c *ConsumerMessageCarrier) Set(key, val string) {
	// Ensure uniqueness of keys
	for i := 0; i < len(c.msg.Headers); i++ {
		if c.msg.Headers[i] != nil && string(c.msg.Headers[i].Key) == key {
			c.msg.Headers = append(c.msg.Headers[:i], c.msg.Headers[i+1:]...)
			i--
		}
	}
	c.msg.Headers = append(
		c.msg.Headers, &sarama.RecordHeader{
			Key:   []byte(key),
			Value: []byte(val),
		},
	)
}

// Keys returns a slice of all key identifiers in the carrier.
func (c *ConsumerMessageCarrier) Keys() []string {
	out := make([]string, len(c.msg.Headers))
	for i, h := range c.msg.Headers {
		out[i] = string(h.Key)
	}
	return out
}

// LoggingInterceptor implements Sarama's ProducerInterceptor and ConsumerInterceptor interfaces.
// It extracts trace context from message headers and logs with trace correlation.
type LoggingInterceptor struct{}

var (
	_ sarama.ProducerInterceptor = (*LoggingInterceptor)(nil)
	_ sarama.ConsumerInterceptor = (*LoggingInterceptor)(nil)
)

// NewLoggingInterceptor creates a new LoggingInterceptor.
func NewLoggingInterceptor() *LoggingInterceptor {
	return &LoggingInterceptor{}
}

// OnSend is called by Sarama just before the message is dispatched to the broker.
// NOTE: msg.Partition and msg.Offset are NOT yet assigned at this point —
// they are always 0. Use the return values of SyncProducer.SendMessage to
// obtain the real partition and offset after the broker acknowledges the write.
func (i *LoggingInterceptor) OnSend(msg *sarama.ProducerMessage) {
	carrier := NewProducerMessageCarrier(msg)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

	var messageValue, messageKey []byte
	if msg.Value != nil {
		messageValue, _ = msg.Value.Encode()
	}
	if msg.Key != nil {
		messageKey, _ = msg.Key.Encode()
	}

	log.For(ctx).Info("Sending message",
		zap.String("topic", msg.Topic),
		zap.ByteString("key", messageKey),
		zap.ByteString("value", messageValue),
	)
}

// OnConsume is called when a message is consumed.
// It extracts trace context from message headers and logs the consume operation.
func (i *LoggingInterceptor) OnConsume(msg *sarama.ConsumerMessage) {
	carrier := NewConsumerMessageCarrier(msg)
	ctx := otel.GetTextMapPropagator().Extract(context.Background(), carrier)

	log.For(ctx).Info("Consumer received message",
		zap.String("topic", msg.Topic),
		zap.Int32("partition", msg.Partition),
		zap.Int64("offset", msg.Offset),
		zap.ByteString("key", msg.Key),
		zap.ByteString("value", msg.Value),
	)
}
