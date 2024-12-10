package kafka

import (
	"strings"

	"github.com/IBM/sarama"

	"github.com/trinhdaiphuc/go-kit/collection"
)

// Config is the configuration for Kafka
type Config struct {
	Brokers   string   `json:"brokers" yaml:"brokers" mapstructure:"brokers"`
	GroupID   string   `json:"group_id" yaml:"group_id" mapstructure:"group_id"`
	ClientID  string   `json:"client_id" yaml:"client_id" mapstructure:"client_id"`
	Username  string   `json:"username" yaml:"username" mapstructure:"username"`
	Password  string   `json:"password" yaml:"password" mapstructure:"password"`
	Topics    []string `json:"topics" yaml:"topics" mapstructure:"topics"`
	Algorithm string   `json:"algorithm" yaml:"algorithm" mapstructure:"algorithm"`
	UseSSL    bool     `json:"use_ssl" yaml:"use_ssl" mapstructure:"use_ssl"`
	VerifySSL bool     `json:"verify_ssl" yaml:"verify_ssl" mapstructure:"verify_ssl"`
	CertFile  string   `json:"cert_file" yaml:"cert_file" mapstructure:"cert_file"`
	KeyFile   string   `json:"key_file" yaml:"key_file" mapstructure:"key_file"`
	CAFile    string   `json:"ca_file" yaml:"ca_file" mapstructure:"ca_file"`
	Retry     *int     `json:"retry" yaml:"retry" mapstructure:"retry"`
	RetryTime int      `json:"retry_time" yaml:"retry_time" mapstructure:"retry_time"`
}

func (c *Config) BrokersArray() []string {
	return strings.Split(c.Brokers, ",")
}

func GetMessagesTopic(messages []*sarama.ProducerMessage) string {
	topics := collection.ToArrayString(
		messages, func(msg *sarama.ProducerMessage) string {
			return msg.Topic
		},
	)

	topics = collection.DeDuplicate(topics)
	result := strings.Join(topics, ",")
	return result
}
