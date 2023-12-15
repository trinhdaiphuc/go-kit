package kafka

import (
	"crypto/sha256"
	"crypto/sha512"
	"hash"

	"github.com/IBM/sarama"
	"github.com/xdg-go/scram"
)

// Algorithm determines the hash function used by SCRAM to protect the user's
// credentials.
type Algorithm interface {
	// Name returns the algorithm's name, e.g. "SCRAM-SHA-256"
	Name() sarama.SASLMechanism

	// Hash returns a new hash.Hash.
	Hash() hash.Hash
}

type sha256Algo struct{}

func (sha256Algo) Name() sarama.SASLMechanism {
	return sarama.SASLTypeSCRAMSHA256
}

func (sha256Algo) Hash() hash.Hash {
	return sha256.New()
}

type sha512Algo struct{}

func (sha512Algo) Name() sarama.SASLMechanism {
	return sarama.SASLTypeSCRAMSHA512
}

func (sha512Algo) Hash() hash.Hash {
	return sha512.New()
}

const (
	// SASLTypeOAuth represents the SASL/OAUTHBEARER mechanism (Kafka 2.0.0+)
	SASLTypeOAuth = "OAUTHBEARER"
	// SASLTypePlaintext represents the SASL/PLAIN mechanism
	SASLTypePlaintext = "PLAIN"
	// SASLTypeSCRAMSHA256 represents the SCRAM-SHA-256 mechanism.
	SASLTypeSCRAMSHA256 = "SCRAM-SHA-256"
	// SASLTypeSCRAMSHA512 represents the SCRAM-SHA-512 mechanism.
	SASLTypeSCRAMSHA512 = "SCRAM-SHA-512"
	SASLTypeGSSAPI      = "GSSAPI"
	// SASLHandshakeV0 is v0 of the Kafka SASL handshake protocol. Client and
	// server negotiate SASL auth using opaque packets.
	SASLHandshakeV0 = int16(0)
	// SASLHandshakeV1 is v1 of the Kafka SASL handshake protocol. Client and
	// server negotiate SASL by wrapping tokens with Kafka protocol headers.
	SASLHandshakeV1 = int16(1)
	// SASLExtKeyAuth is the reserved extension key name sent as part of the
	// SASL/OAUTHBEARER initial client response
	SASLExtKeyAuth = "auth"
)

var (
	SHA256 Algorithm = sha256Algo{}
	SHA512 Algorithm = sha512Algo{}
)

type XDGSCRAMClient struct {
	*scram.Client
	*scram.ClientConversation
	scram.HashGeneratorFcn
}

func NewSCRAMClient(algo Algorithm) func() sarama.SCRAMClient {
	return func() sarama.SCRAMClient {
		return &XDGSCRAMClient{HashGeneratorFcn: algo.Hash}
	}
}

func (x *XDGSCRAMClient) Begin(userName, password, authzID string) error {
	client, err := x.HashGeneratorFcn.NewClient(userName, password, authzID)
	if err != nil {
		return err
	}
	x.Client = client
	x.ClientConversation = x.Client.NewConversation()
	return nil
}

func (x *XDGSCRAMClient) Step(challenge string) (string, error) {
	return x.ClientConversation.Step(challenge)
}

func (x *XDGSCRAMClient) Done() bool {
	return x.ClientConversation.Done()
}
