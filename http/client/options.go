package httpclient

import (
	"net/url"
	"time"

	httptripperware "github.com/trinhdaiphuc/go-kit/http/tripperware"
	tripperwareretry "github.com/trinhdaiphuc/go-kit/http/tripperware/retry"
)

type Options struct {
	proxyURL            *url.URL
	maxConnsPerHost     int
	maxIdleConns        int
	maxIdleConnsPerHost int
	keepAliveTimeout    time.Duration
	requestTimeout      time.Duration
	idleConnTimeout     time.Duration
	retry               []tripperwareretry.Option
	tripperwares        []httptripperware.Tripperware
	isEnablePrometheus  bool
}

type Option func(*Options)

func WithProxyURL(proxyURL string) Option {
	return func(o *Options) {
		proxyURL, err := url.Parse(proxyURL)
		if err != nil {
			panic(err)
		}
		o.proxyURL = proxyURL
	}
}

func WithMaxConnsPerHost(max int) Option {
	return func(o *Options) {
		o.maxConnsPerHost = max
	}
}

func WithMaxIdleConns(max int) Option {
	return func(o *Options) {
		o.maxIdleConns = max
	}
}

func WithMaxIdleConnsPerHost(max int) Option {
	return func(o *Options) {
		o.maxIdleConnsPerHost = max
	}
}

func WithRequestTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.requestTimeout = timeout
	}
}

func WithIdleConnTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.idleConnTimeout = timeout
	}
}

func WithKeepAliveTimeout(timeout time.Duration) Option {
	return func(o *Options) {
		o.keepAliveTimeout = timeout
	}
}

func WithEnablePrometheus(enable bool) Option {
	return func(o *Options) {
		o.isEnablePrometheus = enable
	}
}

func WithRetryOptions(options ...tripperwareretry.Option) Option {
	return func(o *Options) {
		o.retry = options
	}
}

func WithTripperwares(tripperwares ...httptripperware.Tripperware) Option {
	return func(o *Options) {
		o.tripperwares = tripperwares
	}
}
