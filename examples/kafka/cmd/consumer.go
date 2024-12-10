package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/IBM/sarama"

	"github.com/spf13/cobra"

	"github.com/trinhdaiphuc/go-kit/kafka"
	"github.com/trinhdaiphuc/go-kit/metrics"
	"github.com/trinhdaiphuc/go-kit/thread"
	"github.com/trinhdaiphuc/go-kit/tracing"
)

var (
	consumerCmd = &cobra.Command{
		Use:   "consumer",
		Short: "Consumer is a Kafka consumer application.",
		Long:  "Consumer is a Kafka consumer application. It is used to consume messages from Kafka topic.",
		Run: func(cmd *cobra.Command, args []string) {
			consumerRun(cmd, args)
		},
	}
)

func NewConfig() *kafka.Config {
	return &kafka.Config{
		ClientID: clientID,
		Brokers:  brokers,
		GroupID:  groupID,
		Topics:   []string{topic},
	}
}

func NewConsumer() (kafka.Consumer, error) {
	handlerFn := handler
	if useMetrics {
		fmt.Println("Using metrics")
		metrics.NewServerMonitor("kafka-consumer")
		metrics.KafkaConsumerHandlerInterceptor(handlerFn)
	}

	if useTracing {
		fmt.Println("Using tracing")
		_, shutdown, err := tracing.TracerProvider("kafka-consumer", "1.0.0")
		if err != nil {
			panic(err)
			return nil, err
		}
		fmt.Println("Init tracing success")
		defer shutdown()
		handlerFn = tracing.WrapConsumerHandler(handlerFn)
	}

	consumer, err := kafka.NewConsumer(NewConfig(), handlerFn)
	if err != nil {
		return nil, err
	}

	return consumer, nil
}

func consumerRun(_ *cobra.Command, _ []string) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	rg := thread.NewRoutineGroup()

	consumer, err := NewConsumer()
	if err != nil {
		panic(err)
	}

	rg.Run(consumer.Start)

	<-stop

	fmt.Println("Shutting down consumer...")
	rg.Run(consumer.Close)
	rg.Wait()
	fmt.Println("consumer closed")
}

func handler(ctx context.Context, message *sarama.ConsumerMessage) error {
	fmt.Printf("Received message from topic: %s, partition: %d, offset: %d\nValue: %s\n\n", message.Topic, message.Partition, message.Offset, message.Value)
	return nil
}
