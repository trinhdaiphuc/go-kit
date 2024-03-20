package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:   "kafka",
		Short: "A Kafka example application.",
		Long:  `Kafka is a distributed streaming platform that is used to build real-time data pipelines and streaming applications.`,
	}

	clientID, groupID, topic, message string
	brokers                           []string
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&clientID, "client-id", "c", "client-1", "Client ID")
	rootCmd.PersistentFlags().StringVarP(&groupID, "group-id", "g", "group-1", "Group ID")
	rootCmd.PersistentFlags().StringVarP(&topic, "topic", "t", "topic-1", "Topic")
	rootCmd.PersistentFlags().StringVarP(&message, "message", "m", "Hello, Kafka!", "Producer's message")
	rootCmd.PersistentFlags().StringSliceVarP(&brokers, "brokers", "b", []string{"localhost:9092"}, "Kafka brokers address")

	rootCmd.AddCommand(producerCmd)
	rootCmd.AddCommand(consumerCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
