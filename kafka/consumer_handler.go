package kafka

import (
	"github.com/IBM/sarama"
	"golang.org/x/net/context"

	"github.com/trinhdaiphuc/go-kit/log"
)

//go:generate mockgen -destination=./mocks/$GOFILE -source=$GOFILE -package=kafkamock

// ConsumerHandler represents a Sarama consumer group consumer
type ConsumerHandler struct {
	handler ConsumerHandlerFn
}

func NewConsumerHandler(handler ConsumerHandlerFn) sarama.ConsumerGroupHandler {
	return &ConsumerHandler{handler: handler}
}

// ConsumerHandlerFn is invoked for each message received by consumer
type ConsumerHandlerFn func(ctx context.Context, message *sarama.ConsumerMessage) error

// Setup is run at the beginning of a new session, before ConsumeClaim
func (c *ConsumerHandler) Setup(sarama.ConsumerGroupSession) error {
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (c *ConsumerHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (c *ConsumerHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine, see:
	// https://github.com/Shopify/sarama/blob/main/consumer_group.go#L27-L29
	for {
		select {
		case message, ok := <-claim.Messages():
			if !ok { // check message to prevent panic (ISSUE: https://github.com/IBM/sarama/issues/2477)
				log.Bg().Warn("Message channel was closed", log.String("topic", claim.Topic()),
					log.Int32("partition", claim.Partition()), log.Int64("next_offset", claim.HighWaterMarkOffset()))
				return nil
			}

			err := c.handler(session.Context(), message)
			if err != nil {
				return err
			}
			session.MarkMessage(message, "")

		// Should return when `session.Context()` is done.
		// If not, will raise `ErrRebalanceInProgress` or `read tcp <ip>:<port>: i/o timeout` when kafka rebalance. see:
		// https://github.com/Shopify/sarama/issues/1192
		case <-session.Context().Done():
			return nil
		}
	}
}
