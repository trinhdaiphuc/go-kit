package cacheredis

import (
	"errors"
	"fmt"

	"github.com/goccy/go-json"
	"google.golang.org/protobuf/proto"

	"github.com/trinhdaiphuc/go-kit/cache"
)

type Option[K comparable, V any] func(*Options[K, V])

type Marshaller func(v any) ([]byte, error)
type Unmarshaler func(data []byte, v any) error
type KeyEncoder func(key any) string
type KeyDecoder func(key string) any

type Options[K comparable, V any] struct {
	Prefix         string
	KeyEncoder     KeyEncoder
	KeyDecoder     KeyDecoder
	MarshalValue   Marshaller
	UnmarshalValue Unmarshaler
	Loader         cache.Loader[K, V]
}

func newDefaultOption[K comparable, V any]() *Options[K, V] {
	return &Options[K, V]{
		KeyEncoder:     defaultKeyEncoder,
		KeyDecoder:     defaultKeyDecoder,
		MarshalValue:   json.Marshal,
		UnmarshalValue: json.Unmarshal,
	}
}

func WithMarshaller[K comparable, V any](marshaller Marshaller) Option[K, V] {
	return func(o *Options[K, V]) {
		o.MarshalValue = marshaller
	}
}

func WithUnMarshaller[K comparable, V any](unmarshaler Unmarshaler) Option[K, V] {
	return func(o *Options[K, V]) {
		o.UnmarshalValue = unmarshaler
	}
}

func WithKeyEncoder[K comparable, V any](encoder KeyEncoder) Option[K, V] {
	return func(o *Options[K, V]) {
		o.KeyEncoder = encoder
	}
}

func WithKeyDecoder[K comparable, V any](decoder KeyDecoder) Option[K, V] {
	return func(o *Options[K, V]) {
		o.KeyDecoder = decoder
	}
}

func WithLoader[K comparable, V any](loader cache.Loader[K, V]) Option[K, V] {
	return func(o *Options[K, V]) {
		if loader != nil {
			o.Loader = loader
		}
	}
}

func WithPrefix[K comparable, V any](prefix string) Option[K, V] {
	return func(o *Options[K, V]) {
		o.Prefix = prefix
	}
}

func defaultKeyEncoder(key any) string {
	return fmt.Sprint(key)
}

func defaultKeyDecoder(key string) (result any) {
	return key
}

func ProtoMarshaller(v interface{}) ([]byte, error) {
	msg, ok := v.(proto.Message)
	if !ok {
		return nil, errors.New("value is not a proto.Message")
	}
	return proto.Marshal(msg)
}

func ProtoUnmarshaler(data []byte, v interface{}) error {
	msg, ok := v.(proto.Message)
	if !ok {
		return errors.New("value is not a proto.Message")
	}
	return proto.Unmarshal(data, msg)
}
