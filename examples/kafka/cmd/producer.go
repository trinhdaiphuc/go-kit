package cmd

import (
	"fmt"
	"time"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"

	"github.com/trinhdaiphuc/go-kit/kafka"
)

var (
	producerCmd = &cobra.Command{
		Use:   "producer",
		Short: "Producer is a Kafka producer application.",
		Long:  "Producer is a Kafka producer application. It is used to produce messages to Kafka topic.",
		Run: func(cmd *cobra.Command, args []string) {
			producerRun(cmd, args)
		},
	}
)

func NewProducer() (kafka.Producer, error) {
	cli, err := NewClient()

	if err != nil {
		return nil, err
	}

	producer, err := kafka.NewProducer(cli, []string{topic})
	if err != nil {
		return nil, err
	}

	return producer, nil
}

func producerRun(_ *cobra.Command, _ []string) {
	producer, err := NewProducer()
	if err != nil {
		return
	}

	partition, offset, err := producer.SendMessage(&sarama.ProducerMessage{
		Topic:     topic,
		Value:     sarama.ByteEncoder(message),
		Timestamp: time.Time{},
	})
	if err != nil {
		fmt.Printf("error sending message: %v, partition: %d, offset: %d\n", err, partition, offset)
		return
	}

	fmt.Printf("message sent to partition: %d, offset: %d\n", partition, offset)
}
