package kafka

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/IBM/sarama"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/log"
)

type consumer struct {
	cli             sarama.ConsumerGroup
	consumerHandler sarama.ConsumerGroupHandler
	cfg             *Config
	stop            chan bool
	quit            *sync.WaitGroup
}

//go:generate mockgen -destination=./mocks/$GOFILE -source=$GOFILE -package=kafkamock
type Consumer interface {
	Start()
	Close()
}

func NewConsumerClient(cfg *Config, client Client, handler ConsumerHandlerFn) (Consumer, error) {
	cli, err := sarama.NewConsumerGroupFromClient(cfg.GroupID, client)
	if err != nil {
		return nil, fmt.Errorf("error creating the consumer client: %w", err)
	}

	return &consumer{
		cli:             cli,
		cfg:             cfg,
		consumerHandler: NewConsumerHandler(handler),
		stop:            make(chan bool),
		quit:            &sync.WaitGroup{},
	}, nil
}

func NewConsumer(cfg *Config, handler ConsumerHandlerFn) (Consumer, error) {
	client, err := NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("error creating the consumer client: %w", err)
	}

	return NewConsumerClient(cfg, client, handler)
}

func (consumer *consumer) Start() {
	log.Bg().Info("Starting a new kafka consumer")
	consumer.quit.Add(1)
	defer consumer.quit.Done()

	ctx, cancel := context.WithCancel(context.Background())

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			// `Consume` should be called inside an infinite loop, when a
			// server-side rebalance happens, the consumer session will need to be
			// recreated to get the new claims
			if err := consumer.cli.Consume(ctx, consumer.cfg.Topics, consumer.consumerHandler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Bg().Error("Error from consumer", zap.Error(err))
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-consumer.stop
	log.Bg().Info("terminating: via signal")

	cancel()
	wg.Wait()
	if err := consumer.cli.Close(); err != nil {
		log.Bg().Error("Error closing client", zap.Error(err))
	}
}

func (consumer *consumer) Close() {
	close(consumer.stop)
	consumer.quit.Wait()
	log.Bg().Info("Consumer has stopped")
}
