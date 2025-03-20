package metrics

import (
	"context"
	"time"

	"github.com/IBM/sarama"

	"github.com/trinhdaiphuc/go-kit/kafka"
)

type kafkaProducer struct {
	kafka.Producer
}

func (producer *kafkaProducer) SendMessage(msg *sarama.ProducerMessage) (partition int32, offset int64, err error) {
	startTime := time.Now()
	partition, offset, err = producer.Producer.SendMessage(msg)
	elapsedTime := time.Since(startTime).Seconds()

	statusCode := "200" // Success
	if err != nil {
		statusCode = "500" // Error
	}

	doneHTTPHandleRequest(OutboundCall, producerLabelMethod, msg.Topic, statusCode, elapsedTime)

	return
}

func (producer *kafkaProducer) SendMessages(msg []*sarama.ProducerMessage) error {
	startTime := time.Now()
	err := producer.Producer.SendMessages(msg)
	elapsedTime := time.Since(startTime).Seconds()

	statusCode := "200" // Success
	if err != nil {
		statusCode = "500" // Error
	}

	doneHTTPHandleRequest(
		OutboundCall, producerLabelMethod, kafka.GetMessagesTopic(msg), statusCode, elapsedTime,
	)

	return err
}

func NewWrapKafkaProducer(producer kafka.Producer) kafka.Producer {
	return &kafkaProducer{Producer: producer}
}

func KafkaConsumerHandlerInterceptor(handler kafka.ConsumerHandlerFn) kafka.ConsumerHandlerFn {
	return func(ctx context.Context, message *sarama.ConsumerMessage) error {
		startTime := time.Now()
		err := handler(ctx, message)
		elapsedTime := time.Since(startTime).Seconds()

		statusCode := "200" // Success
		if err != nil {
			statusCode = "500" // Error
		}

		doneHTTPHandleRequest(InboundCall, consumerLabelMethod, message.Topic, statusCode, elapsedTime)

		return err
	}
}
