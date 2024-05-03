package cmd

import (
	"bufio"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/IBM/sarama"
	"github.com/spf13/cobra"

	"github.com/trinhdaiphuc/go-kit/kafka"
	"github.com/trinhdaiphuc/go-kit/metrics"
	"github.com/trinhdaiphuc/go-kit/tracing"
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

	if useMetrics {
		fmt.Println("Using metrics")
		metrics.NewServerMonitor("kafka-producer")
		producer = metrics.NewWrapKafkaProducer(producer)
	}

	if useTracing {
		fmt.Println("Using tracing")
		_, shutdown, err := tracing.TracerProvider("kafka-producer", "1.0.0")
		if err != nil {
			panic(err)
			return nil, err
		}
		defer shutdown()
		producer = tracing.WrapKafkaProducer(cli.Config(), producer)
	}

	return producer, nil
}

func producerRun(_ *cobra.Command, _ []string) {
	producer, err := NewProducer()
	if err != nil {
		panic(err)
	}

	sendMessage(producer)

	err = producer.Close()
	if err != nil {
		fmt.Println("error closing producer: ", err)
	}

	fmt.Println("producer closed")
}

func sendMessage(producer kafka.Producer) {
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	reader := bufio.NewReader(os.Stdin)

	for {
		select {
		case <-stop:
			return
		default:
			message, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println("error reading message: ", err)
				return
			}

			// remove newline character
			message = strings.Replace(message, "\n", "", -1)
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

	}
}
