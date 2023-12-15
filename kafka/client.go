package kafka

import (
	"fmt"

	"github.com/IBM/sarama"
)

type client struct {
	sarama.Client
}

type Client interface {
	sarama.Client
	Ping() error
}

// NewClient creates a new Client. It connects to one of the given broker addresses
// and uses that broker to automatically fetch metadata on the rest of the kafka cluster. If metadata cannot
// be retrieved from any of the given broker addresses, the client is not created.
func NewClient(brokers []string, opts ...Option) (Client, error) {
	config := DefaultConfig()
	for _, f := range opts {
		f(config)
	}

	cli, err := sarama.NewClient(brokers, config)
	if err != nil {
		return nil, fmt.Errorf("error creating the kafka client: %w", err)
	}

	return &client{
		Client: cli,
	}, nil
}

// Ping verifies a connection to the kafka is still alive, establishing a connection if necessary.
func (c *client) Ping() error {
	broker, err := c.RefreshController()
	if err != nil {
		return err
	}
	if broker == nil {
		return fmt.Errorf("no broker found")
	}
	connected, err := broker.Connected()
	if err != nil {
		return err
	}
	if !connected {
		return fmt.Errorf("broker not connected")
	}
	return nil
}
