package metrics

import (
	"testing"

	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

// Every collector the Monitor writes to must reach the registry, otherwise the
// series silently never appears on /metrics.
func TestNewServerMonitor_RegistersEveryCollector(t *testing.T) {
	m := NewServerMonitor("test-service")

	collectors := map[string]prom.Collector{
		"requestRates":          m.requestRates,
		"durationSeconds":       m.durationSeconds,
		"clientRequestRates":    m.clientRequestRates,
		"clientDurationSeconds": m.clientDurationSeconds,
		"circuitBreakerState":   m.circuitBreakerState,
		"requestCounter":        m.requestCounter,
		"successCounter":        m.successCounter,
		"failureCounter":        m.failureCounter,
	}

	for name, c := range collectors {
		t.Run(name, func(t *testing.T) {
			// Registering an already-registered collector fails with
			// AlreadyRegisteredError; anything else means it was never registered.
			err := prom.Register(c)
			assert.ErrorAs(t, err, &prom.AlreadyRegisteredError{}, "%s is not registered", name)
		})
	}
}
