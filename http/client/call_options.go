package httpclient

import (
	"net/http"

	"github.com/trinhdaiphuc/go-kit/header"
)

type RequestOption func(req *http.Request)

func WithRequestHeaderKeyValue(key string, value string) RequestOption {
	return func(req *http.Request) {
		req.Header.Add(key, value)
	}
}

func WithRequestHeader(headers header.Headers) RequestOption {
	return func(req *http.Request) {
		for key, value := range headers {
			if value != "" {
				req.Header.Add(key, value)
			}
		}
	}
}

func WithRequestPattern(pattern string) RequestOption {
	return func(req *http.Request) {
		req.Pattern = pattern
	}
}
