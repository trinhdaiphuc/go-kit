package kafka

import (
	"context"
	"time"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/log"
)

// ConsumerBatchHandler represents a Sarama consumer group consumer that buffers
// messages and invokes the handler either when the batch reaches batchSize or
// when delayInterval elapses with pending messages.
type ConsumerBatchHandler struct {
	handler       ConsumerBatchHandlerFn
	batchSize     int
	delayInterval time.Duration
}

// MessageResult holds the processing outcome for a single message.
type MessageResult struct {
	Offset int64
	Error  error
}

const (
	// DefaultBatchSize is equal to the default ChannelBufferSize of Sarama.
	// Do not set batchSize larger than ChannelBufferSize or the channel will block.
	DefaultBatchSize     = 256
	DefaultDelayInterval = 100 * time.Millisecond
)

// NewConsumerBatchHandler creates a new Sarama consumer group handler with batch processing.
// The handler is called when the accumulated batch reaches batchSize OR when
// delayInterval elapses and there are pending messages — whichever comes first.
//
//   - batchSize: max number of messages per batch. Must be <= ChannelBufferSize of Sarama.
//   - delayInterval: max time to wait before flushing a partial batch.
func NewConsumerBatchHandler(handler ConsumerBatchHandlerFn, batchSize int, delayInterval time.Duration) sarama.ConsumerGroupHandler {
	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}
	if delayInterval <= 0 {
		delayInterval = DefaultDelayInterval
	}
	return &ConsumerBatchHandler{
		handler:       handler,
		batchSize:     batchSize,
		delayInterval: delayInterval,
	}
}

// ConsumerBatchHandlerFn is invoked for each batch of messages.
// The returned []MessageResult must have the same length as messages.
// Mark the message as committed only when its corresponding result has a nil Error.
type ConsumerBatchHandlerFn func(ctx context.Context, messages []*sarama.ConsumerMessage) []MessageResult

// Setup is run at the beginning of a new session, before ConsumeClaim.
func (c *ConsumerBatchHandler) Setup(session sarama.ConsumerGroupSession) error {
	claims := session.Claims()
	for topic, partitions := range claims {
		log.Bg().Info("Assigned partitions", zap.String("topic", topic), zap.Int32s("partitions", partitions))
	}
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited.
func (c *ConsumerBatchHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim starts a consumer loop for a single partition claim.
// It buffers messages and flushes the batch when:
//   - the buffer reaches batchSize, or
//   - delayInterval elapses with at least one pending message.
//
// NOTE: Do not move this to a goroutine — ConsumeClaim itself is called within
// a goroutine by Sarama. See https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
func (c *ConsumerBatchHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	log.For(session.Context()).Info("Starting to consume partition",
		zap.String("topic", claim.Topic()),
		zap.Int32("partition", claim.Partition()),
		zap.Int64("initial_offset", claim.InitialOffset()),
		zap.Int64("high_water_mark", claim.HighWaterMarkOffset()))

	batch := make([]*sarama.ConsumerMessage, 0, c.batchSize)
	ticker := time.NewTicker(c.delayInterval)
	defer ticker.Stop()

	flush := func() error {
		if len(batch) == 0 {
			return nil
		}

		results := c.handler(session.Context(), batch)

		for i, msg := range batch {
			if i < len(results) && results[i].Error != nil {
				log.For(session.Context()).Error("Failed to process message",
					zap.String("topic", msg.Topic),
					zap.Int32("partition", msg.Partition),
					zap.Int64("offset", msg.Offset),
					zap.Error(results[i].Error))
				continue
			}
			session.MarkMessage(msg, "")
		}

		batch = batch[:0] // reset, keep allocated capacity
		return nil
	}

	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok {
				// Channel closed — flush remaining messages then exit.
				// https://github.com/IBM/sarama/issues/2477
				log.For(session.Context()).Warn("Message channel was closed",
					zap.String("topic", claim.Topic()),
					zap.Int32("partition", claim.Partition()),
					zap.Int64("next_offset", claim.HighWaterMarkOffset()))
				_ = flush()
				return nil
			}

			batch = append(batch, message)

			if len(batch) >= c.batchSize {
				if err := flush(); err != nil {
					return err
				}
				// Reset ticker so the interval always measures from the last flush.
				ticker.Reset(c.delayInterval)
			}

		case <-ticker.C:
			if err := flush(); err != nil {
				return err
			}

		// Should return when `session.Context()` is done to avoid
		// `ErrRebalanceInProgress` or i/o timeout on Kafka rebalance.
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			_ = flush()
			return nil
		}
	}
}
