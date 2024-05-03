package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kafka",
		Short: "A Kafka example application with producer and consumer commands.",
		Long:  `A Kafka example application with producer and consumer commands. It is used to produce and consume messages from Kafka topic.`,
	}

	clientID, groupID, topic string
	useTracing, useMetrics   bool
	brokers                  []string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&clientID, "client-id", "c", "client-1", "Client ID")
	rootCmd.PersistentFlags().StringVarP(&groupID, "group-id", "g", "group-1", "Group ID")
	rootCmd.PersistentFlags().StringVarP(&topic, "topic", "t", "topic-1", "Topic")
	rootCmd.PersistentFlags().StringSliceVarP(&brokers, "brokers", "b", []string{"localhost:9092"}, "Kafka brokers address")
	rootCmd.PersistentFlags().BoolVarP(&useTracing, "tracing", "r", false, "Enable tracing")
	rootCmd.PersistentFlags().BoolVarP(&useMetrics, "metrics", "e", false, "Enable metrics")

	rootCmd.AddCommand(producerCmd)
	rootCmd.AddCommand(consumerCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
