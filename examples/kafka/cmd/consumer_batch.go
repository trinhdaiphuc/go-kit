package cmd

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/kafka"
	"github.com/trinhdaiphuc/go-kit/log"
	"github.com/trinhdaiphuc/go-kit/thread"
)

var (
	batchSize     int
	batchInterval time.Duration

	batchConsumerCmd = &cobra.Command{
		Use:   "batch-consumer",
		Short: "Batch consumer — processes messages in batches.",
		Long: `batch-consumer buffers incoming Kafka messages and invokes the handler
when either the batch reaches --batch-size or --batch-interval elapses,
whichever comes first.`,
		Run: func(cmd *cobra.Command, args []string) {
			batchConsumerRun(cmd, args)
		},
	}
)

func init() {
	batchConsumerCmd.Flags().IntVarP(&batchSize, "batch-size", "s", kafka.DefaultBatchSize,
		"Max number of messages per batch (must be <= Sarama ChannelBufferSize, default 256)")
	batchConsumerCmd.Flags().DurationVarP(&batchInterval, "batch-interval", "n", kafka.DefaultDelayInterval,
		"Max time to wait before flushing a partial batch (e.g. 200ms, 1s)")

	rootCmd.AddCommand(batchConsumerCmd)
}

// NewBatchConsumer wires up a kafka.Consumer backed by ConsumerBatchHandler.
func NewBatchConsumer() (kafka.Consumer, error) {
	cfg := NewConfig()

	log.Bg().Info("Starting batch consumer",
		zap.Reflect("config", cfg),
		zap.Int("batch_size", batchSize),
		zap.Duration("batch_interval", batchInterval))

	client, err := kafka.NewClient(cfg, kafka.WithConsumerInterceptor(kafka.NewLoggingInterceptor()))
	if err != nil {
		return nil, err
	}

	return kafka.NewBatchConsumerClient(cfg, client, batchHandler, batchSize, batchInterval)
}

func batchConsumerRun(_ *cobra.Command, _ []string) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	rg := thread.NewRoutineGroup()

	consumer, err := NewBatchConsumer()
	if err != nil {
		log.Bg().Fatal("Failed to create batch consumer", zap.Error(err))
	}

	rg.Run(consumer.Start)

	<-stop

	log.Bg().Info("Shutting down batch consumer...")
	rg.Run(consumer.Close)
	rg.Wait()
	log.Bg().Info("Batch consumer closed")
}

// batchHandler is invoked once per flush with all buffered messages.
// It must return a []kafka.MessageResult of exactly the same length.
func batchHandler(ctx context.Context, messages []*sarama.ConsumerMessage) []kafka.MessageResult {
	log.For(ctx).Info("Batch received",
		zap.Int("count", len(messages)),
		zap.String("topic", messages[0].Topic),
		zap.Int32("partition", messages[0].Partition),
		zap.Int64("first_offset", messages[0].Offset),
		zap.Int64("last_offset", messages[len(messages)-1].Offset),
	)

	results := make([]kafka.MessageResult, len(messages))

	for i, msg := range messages {
		// Simulate per-message processing work.
		time.Sleep(5 * time.Millisecond)

		log.For(ctx).Debug("Processed message",
			zap.Int("index", i),
			zap.Int64("offset", msg.Offset),
			zap.ByteString("value", msg.Value),
		)

		results[i] = kafka.MessageResult{Offset: msg.Offset, Error: nil}
	}

	log.For(ctx).Info("Batch processed successfully", zap.Int("count", len(messages)))
	return results
}
