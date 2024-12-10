package kafka

import (
	"fmt"

	"github.com/IBM/sarama"

	"github.com/trinhdaiphuc/go-kit/log"
)

type producer struct {
	sarama.SyncProducer
	cli Client
	cfg *Config
}

//go:generate mockgen -destination=./mocks/$GOFILE -source=$GOFILE -package=kafkamock
type Producer interface {
	sarama.SyncProducer
	Topics() []string
	GetClient() Client
}

func NewProducerClient(cfg *Config, client Client) (Producer, func(), error) {
	producerCli, err := sarama.NewSyncProducerFromClient(client)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating the producer client: %w", err)
	}

	cleanup := func() {
		err = client.Close()
		if err != nil {
			log.Bg().Error("Close kafka consumer client failed", log.Error(err))
		} else {
			log.Bg().Info("Close kafka consumer client succeeded")
		}
	}

	return &producer{
		SyncProducer: producerCli,
		cli:          client,
		cfg:          cfg,
	}, cleanup, nil
}

func NewProducer(cfg *Config) (Producer, func(), error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, nil, fmt.Errorf("error creating the producer client: %w", err)
	}

	return NewProducerClient(cfg, client)
}

func (p *producer) Topics() []string {
	return p.cfg.Topics
}

func (p *producer) GetClient() Client {
	return p.cli
}
