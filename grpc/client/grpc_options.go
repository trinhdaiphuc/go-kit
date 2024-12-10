package grpcclient

import (
	"time"

	"google.golang.org/grpc"
)

type Config struct {
	UnaryClientInterceptors  []grpc.UnaryClientInterceptor
	StreamClientInterceptors []grpc.StreamClientInterceptor
	Timeout                  time.Duration
	RetryMaxTimes            uint
	DNSResolver              bool
	TLSTransport             bool
	MonitoringEnabled        bool
}

type Option func(cfg *Config)

const (
	defaultTimeout       = time.Second
	defaultRetryMaxTimes = 1
)

func WithUnaryClientInterceptors(unaryClientInterceptors ...grpc.UnaryClientInterceptor) Option {
	return func(cfg *Config) {
		cfg.UnaryClientInterceptors = unaryClientInterceptors
	}
}

func WithStreamClientInterceptors(streamClientInterceptors ...grpc.StreamClientInterceptor) Option {
	return func(cfg *Config) {
		cfg.StreamClientInterceptors = streamClientInterceptors
	}
}

func WithTLSTransport(tlsTransport bool) Option {
	return func(cfg *Config) {
		cfg.TLSTransport = tlsTransport
	}
}

func WithTimeout(timeout time.Duration) Option {
	return func(cfg *Config) {
		if timeout <= 0 {
			timeout = defaultTimeout
		}
		cfg.Timeout = timeout
	}
}

func WithRetryMaxTimes(retry uint) Option {
	return func(cfg *Config) {
		if retry <= 0 {
			retry = defaultRetryMaxTimes
		}
		cfg.RetryMaxTimes = retry
	}
}

func WithDNSResolver(enable bool) Option {
	return func(cfg *Config) {
		cfg.DNSResolver = enable
	}
}

func WithMonitoring(enable bool) Option {
	return func(cfg *Config) {
		cfg.MonitoringEnabled = enable
	}
}

func defaultConfig() *Config {
	return &Config{
		UnaryClientInterceptors:  nil,
		StreamClientInterceptors: nil,
		Timeout:                  defaultTimeout,
		RetryMaxTimes:            defaultRetryMaxTimes,
		TLSTransport:             false,
		DNSResolver:              true,
		MonitoringEnabled:        true,
	}
}
