package kafka

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"os"
	"time"

	"github.com/IBM/sarama"

	"github.com/trinhdaiphuc/go-kit/log"
)

//go:generate mockgen -destination=./mocks/$GOFILE -source=$GOFILE -package=kafkamock
type client struct {
	sarama.Client
}

type Client interface {
	sarama.Client
	Ping() error
}

func NewClient(cfg *Config, opts ...Option) (Client, error) {
	opts = append([]Option{
		WithClientID(cfg.ClientID),
		WithConsumerGroupBalance(sarama.NewBalanceStrategyRoundRobin()),
	}, opts...)

	if cfg.Username != "" && cfg.Password != "" {
		sasl := SASL{
			Enable:    true,
			Handshake: true,
			Version:   SASLHandshakeV1, // Use SASL v1 for Kafka > 1.x
			User:      cfg.Username,
			Password:  cfg.Password,
		}

		switch cfg.Algorithm {
		case "plain":
			sasl.Mechanism = sarama.SASLTypePlaintext
		case "sha256":
			sasl.SCRAMClientGeneratorFunc = NewSCRAMClient(SHA256)
			sasl.Mechanism = SHA256.Name()
		case "sha512":
			sasl.SCRAMClientGeneratorFunc = NewSCRAMClient(SHA512)
			sasl.Mechanism = SHA512.Name()
		default:
			// Default to PLAIN if algorithm not specified
			sasl.Mechanism = sarama.SASLTypePlaintext
		}
		opts = append(opts, WithSASLClient(sasl))
	}

	if cfg.UseSSL {
		tlsClient, err := createTLSConfiguration(cfg)
		if err != nil {
			log.Bg().Error("create tls configuration failed: ", log.Error(err))
			return nil, err
		}
		tlsConfig := TLS{
			Enable: true,
			Config: tlsClient,
		}
		opts = append(opts, WithTLSClient(tlsConfig))
	}

	if cfg.Retry != nil {
		retry := *cfg.Retry
		retryTime := time.Duration(cfg.RetryTime) * time.Millisecond
		opts = append(
			opts,
			WithProducerRetry(
				ProducerRetry{
					Max:     retry,
					Backoff: retryTime,
				},
			),
			WithConsumerRetry(
				ConsumerRetry{
					Backoff: retryTime,
				},
			),
		)
	}

	if cfg.Idempotent {
		opts = append(opts, WithIdempotentProducer())
	}

	if cfg.Compression != "" {
		opts = append(opts, WithCompression(parseCompression(cfg.Compression)))
	}

	cli, err := newClient(cfg.BrokersArray(), opts...)
	if err != nil {
		log.Bg().Error("New kafka client failed", log.Error(err))
		return nil, err
	}

	return cli, nil
}

func createTLSConfiguration(config *Config) (*tls.Config, error) {
	t := &tls.Config{
		InsecureSkipVerify: config.VerifySSL,
	}
	if config.CertFile != "" && config.KeyFile != "" && config.CAFile != "" {
		cert, err := tls.LoadX509KeyPair(config.CertFile, config.KeyFile)
		if err != nil {
			return nil, err
		}

		caCert, err := os.ReadFile(config.CAFile)
		if err != nil {
			return nil, err
		}

		caCertPool := x509.NewCertPool()
		caCertPool.AppendCertsFromPEM(caCert)

		t.Certificates = []tls.Certificate{cert}
		t.RootCAs = caCertPool
	}
	return t, nil
}

// newClient creates a new Sarama Client. It connects to one of the given broker addresses
// and uses that broker to automatically fetch metadata on the rest of the kafka cluster. If metadata cannot
// be retrieved from any of the given broker addresses, the client is not created.
func newClient(brokers []string, opts ...Option) (Client, error) {
	config := DefaultConfig()
	for _, f := range opts {
		f(config)
	}

	err := config.Validate()
	if err != nil {
		return nil, fmt.Errorf("error validating the kafka config: %w", err)
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

func parseCompression(scheme string) sarama.CompressionCodec {
	switch scheme {
	case "none":
		return sarama.CompressionNone
	case "gzip":
		return sarama.CompressionGZIP
	case "snappy":
		return sarama.CompressionSnappy
	case "lz4":
		return sarama.CompressionLZ4
	default:
		return sarama.CompressionNone
	}
}
