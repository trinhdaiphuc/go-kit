package kafka

import (
	"crypto/tls"
	"time"

	"github.com/IBM/sarama"
	"github.com/rcrowley/go-metrics"
	"golang.org/x/net/proxy"
)

// AdminConfig is the namespace for ClusterAdmin properties used by the administrative Kafka client.
type AdminConfig struct {
	Retry struct {
		// The total number of times to retry sending (retriable) admin requests (default 5).
		// Similar to the `retries` setting of the JVM AdminClientConfig.
		Max int
		// Backoff time between retries of a failed request (default 100ms)
		Backoff time.Duration
	}
	// The maximum duration the administrative Kafka client will wait for ClusterAdmin operations,
	// including topics, brokers, configurations and ACLs (defaults to 3 seconds).
	Timeout time.Duration
}

// SASL based authentication with broker. While there are multiple SASL authentication methods
// the current implementation is limited to plaintext (SASL/PLAIN) authentication
type SASL struct {
	// Whether or not to use SASL authentication when connecting to the broker
	// (defaults to false).
	Enable bool
	// SASLMechanism is the name of the enabled SASL mechanism.
	// Possible values: OAUTHBEARER, PLAIN (defaults to PLAIN).
	Mechanism sarama.SASLMechanism
	// Version is the SASL Protocol Version to use
	// Kafka > 1.x should use V1, except on Azure EventHub which use V0
	Version int16
	// Whether or not to send the Kafka SASL handshake first if enabled
	// (defaults to true). You should only set this to false if you're using
	// a non-Kafka SASL proxy.
	Handshake bool
	// AuthIdentity is an (optional) authorization identity (authzid) to
	// use for SASL/PLAIN authentication (if different from User) when
	// an authenticated user is permitted to act as the presented
	// alternative user. See RFC4616 for details.
	AuthIdentity string
	// User is the authentication identity (authcid) to present for
	// SASL/PLAIN or SASL/SCRAM authentication
	User string
	// Password for SASL/PLAIN authentication
	Password string
	// authz id used for SASL/SCRAM authentication
	SCRAMAuthzID string
	// SCRAMClientGeneratorFunc is a generator of a user provided implementation of a SCRAM
	// client used to perform the SCRAM exchange with the server.
	SCRAMClientGeneratorFunc func() sarama.SCRAMClient
	// TokenProvider is a user-defined callback for generating
	// access tokens for SASL/OAUTHBEARER auth. See the
	// AccessTokenProvider interface docs for proper implementation
	// guidelines.
	TokenProvider sarama.AccessTokenProvider

	GSSAPI sarama.GSSAPIConfig
}

type TLS struct {
	// Whether or not to use TLS when connecting to the broker
	// (defaults to false).
	Enable bool
	// The TLS configuration to use for secure connections if
	// enabled (defaults to nil).
	Config *tls.Config
}

type Proxy struct {
	// Whether or not to use proxy when connecting to the broker
	// (defaults to false).
	Enable bool
	// The proxy dialer to use enabled (defaults to nil).
	Dialer proxy.Dialer
}

type ProducerRetry struct {
	// The total number of times to retry sending a message (default 3).
	// Similar to the `message.send.max.retries` setting of the JVM producer.
	Max int
	// How long to wait for the cluster to settle between retries
	// (default 100ms). Similar to the `retry.backoff.ms` setting of the
	// JVM producer.
	Backoff time.Duration
	// Called to compute backoff time dynamically. Useful for implementing
	// more sophisticated backoff strategies. This takes precedence over
	// `Backoff` if set.
	BackoffFunc func(retries, maxRetries int) time.Duration
}

type ConsumerRetry struct {
	// How long to wait after a failing to read from a partition before
	// trying again (default 2s).
	Backoff time.Duration
	// Called to compute backoff time dynamically. Useful for implementing
	// more sophisticated backoff strategies. This takes precedence over
	// `Backoff` if set.
	BackoffFunc func(retries int) time.Duration
}

// ConsumerOffsets specifies configuration for how and when to commit consumed
// offsets. This currently requires the manual use of an OffsetManager
// but will eventually be automated.
type ConsumerOffsets struct {
	// Deprecated: CommitInterval exists for historical compatibility
	// and should not be used. Please use Consumer.Offsets.AutoCommit
	CommitInterval time.Duration

	// AutoCommit specifies configuration for commit messages automatically.
	AutoCommit struct {
		// Whether or not to auto-commit updated offsets back to the broker.
		// (default enabled).
		Enable bool

		// How frequently to commit updated offsets. Ineffective unless
		// auto-commit is enabled (default 1s)
		Interval time.Duration
	}

	// The initial offset to use if no offset was previously committed.
	// Should be OffsetNewest or OffsetOldest. Defaults to OffsetNewest.
	Initial int64

	// The retention duration for committed offsets. If zero, disabled
	// (in which case the `offsets.retention.minutes` option on the
	// broker will be used).  Kafka only supports precision up to
	// milliseconds; nanoseconds will be truncated. Requires Kafka
	// broker version 0.9.0 or later.
	// (default is 0: disabled).
	Retention time.Duration

	Retry struct {
		// The total number of times to retry failing commit
		// requests during OffsetManager shutdown (default 3).
		Max int
	}
}

const (
	// time sarama-cluster assumes the processing of an event may take
	defaultMaxProcessingTime = 1 * time.Second

	// producer flush configuration
	defaultFlushFrequency = 100 * time.Millisecond
	defaultFlushBytes     = 64 * 1024

	defaultClientID = "bank-mapping"
)

// DefaultConfig creates a new config used per default
func DefaultConfig() *sarama.Config {
	metrics.UseNilMetrics = true

	config := sarama.NewConfig()

	config.Version = sarama.V2_0_0_0

	// consumer configuration
	config.Consumer.Return.Errors = true
	config.Consumer.MaxProcessingTime = defaultMaxProcessingTime
	config.Consumer.Offsets.Initial = sarama.OffsetNewest // this configures the initial offset for streams. Tables are always
	config.Consumer.Group.Rebalance.GroupStrategies = []sarama.BalanceStrategy{sarama.NewBalanceStrategyRange()}

	// producer configuration
	config.Producer.RequiredAcks = sarama.WaitForLocal
	config.Producer.Compression = sarama.CompressionSnappy
	config.Producer.Flush.Frequency = defaultFlushFrequency
	config.Producer.Flush.Bytes = defaultFlushBytes
	config.Producer.Return.Successes = true
	config.Producer.Return.Errors = true
	config.ClientID = defaultClientID

	return config
}
