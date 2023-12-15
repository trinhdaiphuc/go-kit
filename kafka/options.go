package kafka

import (
	"fmt"
	"log"

	"github.com/IBM/sarama"
)

type Option func(*sarama.Config)

func WithVersion(v string) Option {
	version, err := sarama.ParseKafkaVersion(v)
	if err != nil {
		log.Panicf("Error parsing Kafka version: %v", err)
	}
	return func(c *sarama.Config) {
		c.Version = version
	}
}

func WithClientID(clientID string) Option {
	return func(c *sarama.Config) {
		c.ClientID = clientID
	}
}

func WithConsumerGroupBalance(strategy ...sarama.BalanceStrategy) Option {
	return func(c *sarama.Config) {
		c.Consumer.Group.Rebalance.GroupStrategies = strategy
	}
}

func WithAdminConfig(adminConfig AdminConfig) Option {
	return func(c *sarama.Config) {
		c.Admin = adminConfig
	}
}

func WithSASLClient(sasl SASL) Option {
	return func(c *sarama.Config) {
		c.Net.SASL = sasl
	}
}

func WithTLSClient(tls TLS) Option {
	return func(c *sarama.Config) {
		c.Net.TLS = tls
	}
}

func WithProxyClient(proxy Proxy) Option {
	return func(c *sarama.Config) {
		c.Net.Proxy = proxy
	}
}

func WithProducerRetry(retry ProducerRetry) Option {
	return func(c *sarama.Config) {
		c.Producer.Retry = retry
	}
}

// WithProducerPartitioner use partitioner to generates partition to send messages to
// (defaults to hashing the message key). Similar to the `partitioner.class`
// setting for the JVM producer.
func WithProducerPartitioner(partitioner sarama.PartitionerConstructor) Option {
	return func(c *sarama.Config) {
		c.Producer.Partitioner = partitioner
	}
}

func WithConsumerRetry(retry ConsumerRetry) Option {
	return func(c *sarama.Config) {
		c.Consumer.Retry = retry
	}
}

// ParsePartitioner return PartitionerConstructor from scheme. Value: manual, hash, random, round-robin
func ParsePartitioner(scheme string) (partitioner sarama.PartitionerConstructor, err error) {
	switch scheme {
	case "manual":
		partitioner = sarama.NewManualPartitioner
	case "hash":
		partitioner = sarama.NewHashPartitioner
	case "random":
		partitioner = sarama.NewRandomPartitioner
	case "round-robin":
		partitioner = sarama.NewRoundRobinPartitioner
	default:
		err = fmt.Errorf("unknow partitioner %v", scheme)
	}
	return
}

// ParseBalanceStrategy return BalanceStrategy from assignor. Value sticky, round-robin, range
func ParseBalanceStrategy(assignor string) (strategy sarama.BalanceStrategy, err error) {
	switch assignor {
	case sarama.StickyBalanceStrategyName:
		strategy = sarama.NewBalanceStrategySticky()
	case sarama.RoundRobinBalanceStrategyName:
		strategy = sarama.NewBalanceStrategyRoundRobin()
	case sarama.RangeBalanceStrategyName:
		strategy = sarama.NewBalanceStrategyRange()
	default:
		err = fmt.Errorf("unrecognized consumer group balance strategy: %s", assignor)
	}
	return
}

func WithOldestOffset() Option {
	return func(c *sarama.Config) {
		c.Consumer.Offsets.Initial = sarama.OffsetOldest
	}
}

func WithProducerInterceptor(interceptors ...sarama.ProducerInterceptor) Option {
	return func(c *sarama.Config) {
		c.Producer.Interceptors = interceptors
	}
}

func WithConsumerInterceptor(interceptors ...sarama.ConsumerInterceptor) Option {
	return func(c *sarama.Config) {
		c.Consumer.Interceptors = interceptors
	}
}
