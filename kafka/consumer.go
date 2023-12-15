package kafka

import (
	"context"
	"errors"
	"fmt"
	"log"
	"sync"

	"github.com/IBM/sarama"
)

type consumer struct {
	cli             sarama.ConsumerGroup
	consumerHandler sarama.ConsumerGroupHandler
	topics          []string
	stop            chan bool
	quit            *sync.WaitGroup
}

type Consumer interface {
	Start()
	Close()
}

func NewConsumer(client sarama.Client, groupID string, topics []string, handler ConsumerHandlerFn) (Consumer, error) {
	cli, err := sarama.NewConsumerGroupFromClient(groupID, client)
	if err != nil {
		return nil, fmt.Errorf("error creating the consumer client: %w", err)
	}

	return &consumer{
		cli:             cli,
		topics:          topics,
		consumerHandler: NewConsumerHandler(handler),
		stop:            make(chan bool),
		quit:            &sync.WaitGroup{},
	}, nil
}

func (consumer *consumer) Start() {
	log.Println("Starting a new kafka consumer")
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
			if err := consumer.cli.Consume(ctx, consumer.topics, consumer.consumerHandler); err != nil {
				if errors.Is(err, sarama.ErrClosedConsumerGroup) {
					return
				}
				log.Printf("Error from consumer %v\n", err)
			}
			// check if context was cancelled, signaling that the consumer should stop
			if ctx.Err() != nil {
				return
			}
		}
	}()

	<-consumer.stop
	log.Println("terminating: via signal")

	cancel()
	wg.Wait()
	if err := consumer.cli.Close(); err != nil {
		log.Printf("Error closing client err %v", err)
	}
}

func (consumer *consumer) Close() {
	close(consumer.stop)
	consumer.quit.Wait()
	log.Println("Consumer has stopped")
}
