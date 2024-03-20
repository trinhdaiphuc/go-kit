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
	"github.com/trinhdaiphuc/go-kit/thread"
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

func NewClient() (kafka.Client, error) {
	var (
		opts = []kafka.Option{
			kafka.WithClientID(clientID),
			kafka.WithConsumerGroupBalance(sarama.NewBalanceStrategyRoundRobin()),
			kafka.WithProducerPartitioner(sarama.NewRandomPartitioner),
		}
	)

	cli, err := kafka.NewClient(brokers, opts...)
	if err != nil {
		return nil, err
	}

	return cli, nil
}

func NewConsumer() (kafka.Consumer, error) {
	cli, err := NewClient()

	if err != nil {
		return nil, err
	}

	consumer, err := kafka.NewConsumer(cli, groupID, []string{topic}, handler)
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

	rg.Run(consumer.Close)
	rg.Wait()
}

func handler(ctx context.Context, message *sarama.ConsumerMessage) error {
	fmt.Printf("Received message: %s\n, topic: %s, partition: %d, offset: %d\n", message.Value, message.Topic, message.Partition, message.Offset)
	return nil
}
