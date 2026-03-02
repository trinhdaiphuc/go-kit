package httpclient

import (
	"bytes"
	"context"
	"io"
	"net"
	"net/http"
	"time"

	"github.com/goccy/go-json"
	"github.com/google/go-querystring/query"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.uber.org/zap"

	"github.com/trinhdaiphuc/go-kit/header"
	httptripperware "github.com/trinhdaiphuc/go-kit/http/tripperware"
	trippercb "github.com/trinhdaiphuc/go-kit/http/tripperware/circuitbreaker"
	tripperwareretry "github.com/trinhdaiphuc/go-kit/http/tripperware/retry"
	"github.com/trinhdaiphuc/go-kit/log"
	"github.com/trinhdaiphuc/go-kit/metrics"
)

const (
	defaultMaxIdleConns        = 100
	defaultMaxIdleConnsPerHost = 100
	defaultKeepAliveTimeout    = 30 * time.Second
	defaultRequestTimeout      = 30 * time.Second
	defaultIdleConnTimeout     = 30 * time.Minute
)

//go:generate mockgen -destination=./mocks/mock_$GOFILE -source=$GOFILE -package=mocks
type Client interface {
	Get(ctx context.Context, url string, request any, opts ...RequestOption) ([]byte, int, error)
	Post(ctx context.Context, url string, data any, opts ...RequestOption) ([]byte, int, error)
	Put(ctx context.Context, url string, data any, opts ...RequestOption) ([]byte, int, error)
	Delete(ctx context.Context, url string, request any, opts ...RequestOption) ([]byte, int, error)
}

type httpClient struct {
	client *http.Client
}

func New(serviceName string, opts ...Option) Client {
	return &httpClient{
		client: NewClient(serviceName, opts...),
	}
}

func NewClient(serviceName string, opts ...Option) *http.Client {
	options := configure(opts...)

	transport := &http.Transport{
		DialContext: (&net.Dialer{
			KeepAlive: options.keepAliveTimeout,
		}).DialContext,
		MaxIdleConns:        options.maxIdleConns,
		MaxIdleConnsPerHost: options.maxIdleConnsPerHost,
		IdleConnTimeout:     options.idleConnTimeout,
	}

	if options.proxyURL != nil {
		transport.Proxy = http.ProxyURL(options.proxyURL)
	}

	options.tripperwares = append(options.tripperwares, tripperwareretry.Tripperware(options.retry...))
	if options.isEnablePrometheus {
		options.tripperwares = append(options.tripperwares, metrics.ClientHTTPTripperware(metrics.WithServiceName(serviceName)))
	}
	if options.enableCircuitBreaker {
		options.tripperwares = append(options.tripperwares, trippercb.Tripperware(options.circuitBreakerOpts...))
	}

	client := &http.Client{
		Transport: otelhttp.NewTransport(transport, otelhttp.WithServerName(serviceName)),
		Timeout:   options.requestTimeout,
	}
	return httptripperware.WrapClient(client, options.tripperwares...)
}

func (h *httpClient) do(req *http.Request) ([]byte, int, error) {
	var (
		resp *http.Response
		errR error
	)

	resp, errR = h.client.Do(req)
	if errR != nil {
		status := http.StatusInternalServerError
		if resp != nil {
			status = resp.StatusCode
		}
		return nil, status, errR
	}

	body, err := io.ReadAll(resp.Body)
	defer resp.Body.Close()

	return body, resp.StatusCode, err
}

func (h *httpClient) Get(ctx context.Context, url string, request any, opts ...RequestOption) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	qs, err := query.Values(request)
	if err != nil {
		log.For(ctx).Error("Repair request failed", zap.Reflect("request", request), zap.Error(err))
		return nil, http.StatusBadRequest, err
	}

	qr := req.URL.Query()
	for k, values := range qs {
		for _, v := range values {
			qr.Add(k, v)
		}
	}
	req.URL.RawQuery = qr.Encode()
	req.Header.Set(header.HTTPContentType, header.ContentTypeApplicationJSON)

	for _, opt := range opts {
		if opt != nil {
			opt(req)
		}
	}

	return h.do(req)
}

func (h *httpClient) Post(ctx context.Context, url string, data any, opts ...RequestOption) ([]byte, int, error) {
	dataByte, err := json.Marshal(data)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(dataByte))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	req.Header.Set(header.HTTPContentType, header.ContentTypeApplicationJSON)

	for _, opt := range opts {
		if opt != nil {
			opt(req)
		}
	}

	return h.do(req)
}

func (h *httpClient) Put(ctx context.Context, url string, data any, opts ...RequestOption) ([]byte, int, error) {
	dataByte, err := json.Marshal(data)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPut, url, bytes.NewBuffer(dataByte))
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	req.Header.Set(header.HTTPContentType, header.ContentTypeApplicationJSON)

	for _, opt := range opts {
		if opt != nil {
			opt(req)
		}
	}

	return h.do(req)
}

func (h *httpClient) Delete(ctx context.Context, url string, request any, opts ...RequestOption) ([]byte, int, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodDelete, url, nil)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}

	qs, err := query.Values(request)
	if err != nil {
		log.For(ctx).Error("Repair request failed", zap.Reflect("request", request), zap.Error(err))
		return nil, http.StatusBadRequest, err
	}

	qr := req.URL.Query()
	for k, values := range qs {
		for _, v := range values {
			qr.Add(k, v)
		}
	}
	req.URL.RawQuery = qr.Encode()

	req.Header.Set(header.HTTPContentType, header.ContentTypeApplicationJSON)
	for _, opt := range opts {
		if opt != nil {
			opt(req)
		}
	}

	return h.do(req)
}

func configure(opts ...Option) *Options {
	options := &Options{
		requestTimeout:      defaultRequestTimeout,
		keepAliveTimeout:    defaultKeepAliveTimeout,
		maxIdleConns:        defaultMaxIdleConns,
		maxIdleConnsPerHost: defaultMaxIdleConnsPerHost,
		idleConnTimeout:     defaultIdleConnTimeout,
		isEnablePrometheus:  true,
	}

	for _, o := range opts {
		o(options)
	}

	return options
}
