package cmd

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/log"
)

// mockMessages is the pre-defined list of messages that will be published in order.
// Edit this slice to change what the mock producer sends.
var mockMessages = []string{
	`{"id":1,"event":"order.created","amount":100}`,
	`{"id":2,"event":"order.created","amount":250}`,
	`{"id":3,"event":"order.paid","amount":100}`,
	`{"id":4,"event":"order.cancelled","amount":250}`,
	`{"id":5,"event":"order.created","amount":75}`,
	`{"id":6,"event":"order.paid","amount":75}`,
	`{"id":7,"event":"order.created","amount":500}`,
	`{"id":8,"event":"order.created","amount":320}`,
	`{"id":9,"event":"order.paid","amount":500}`,
	`{"id":10,"event":"order.refunded","amount":320}`,
}

var (
	mockCount    int
	mockInterval time.Duration
	mockRepeat   bool

	mockProducerCmd = &cobra.Command{
		Use:   "mock-producer",
		Short: "Mock producer — auto-publishes a pre-defined list of messages.",
		Long: `mock-producer publishes a fixed list of messages to a Kafka topic
without requiring any user input.  Useful for testing the batch consumer.

Flags:
  --count       how many messages to send (default: len(mockMessages))
  --interval    delay between each send  (default: 50ms)
  --repeat      loop the list until Ctrl+C
`,
		Run: func(cmd *cobra.Command, args []string) {
			mockProducerRun(cmd, args)
		},
	}
)

func init() {
	mockProducerCmd.Flags().IntVarP(&mockCount, "count", "N", len(mockMessages),
		"Number of messages to send (cycles through the list if larger than list size)")
	mockProducerCmd.Flags().DurationVarP(&mockInterval, "interval", "d", 50*time.Millisecond,
		"Delay between each message send (e.g. 50ms, 1s)")
	mockProducerCmd.Flags().BoolVarP(&mockRepeat, "repeat", "R", false,
		"Repeat the message list indefinitely until Ctrl+C")

	rootCmd.AddCommand(mockProducerCmd)
}

func mockProducerRun(_ *cobra.Command, _ []string) {
	producer, cleanup, err := NewProducer()
	if err != nil {
		log.Bg().Fatal("Failed to create producer", zap.Error(err))
	}
	defer func() {
		cleanup()
		if err = producer.Close(); err != nil {
			log.Bg().Error("Error closing producer", zap.Error(err))
		}
		log.Bg().Info("Mock producer closed")
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(stop)

	ticker := time.NewTicker(mockInterval)
	defer ticker.Stop()

	sent := 0
	total := mockCount
	if mockRepeat {
		total = -1 // unlimited
	}

	log.Bg().Info("Mock producer started",
		zap.String("topic", topic),
		zap.Int("total", mockCount),
		zap.Duration("interval", mockInterval),
		zap.Bool("repeat", mockRepeat),
	)

	for {
		if total >= 0 && sent >= total {
			log.Bg().Info("Mock producer finished", zap.Int("sent", sent))
			return
		}

		select {
		case <-stop:
			fmt.Println()
			log.Bg().Info("Mock producer stopped by signal", zap.Int("sent", sent))
			return

		case <-ticker.C:
			payload := mockMessages[sent%len(mockMessages)]

			partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
				Topic:     topic,
				Key:       sarama.StringEncoder(fmt.Sprintf("key-%d", sent+1)),
				Value:     sarama.ByteEncoder(payload),
				Timestamp: time.Now(),
			})
			if err != nil {
				log.Bg().Error("Failed to send message",
					zap.Int("seq", sent+1),
					zap.String("payload", payload),
					zap.Error(err))
				return
			}

			log.Bg().Info("Message sent",
				zap.Int("seq", sent+1),
				zap.String("topic", topic),
				zap.Int32("partition", partition),
				zap.Int64("offset", offset),
				zap.String("key", fmt.Sprintf("key-%d", sent+1)),
				zap.String("payload", payload),
			)
			sent++
		}
	}
}
