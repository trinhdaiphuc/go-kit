package metrics

import (
	"net/http"
)

// Filter is a predicate used to determine whether a given http.request should
// be monitoring. A Filter must return true if the request should be monitoring.
type Filter func(*http.Request) bool

type config struct {
	Filters []Filter
}

// Option specifies instrumentation configuration options.
type Option interface {
	apply(*config)
}

type optionFunc func(*config)

func (o optionFunc) apply(c *config) {
	o(c)
}

// WithFilter adds a filter to the list of filters used by the handler.
// If any filter indicates to exclude a request then the request will not be
// monitoring.
func WithFilter(f ...Filter) Option {
	return optionFunc(func(c *config) {
		c.Filters = append(c.Filters, f...)
	})
}
