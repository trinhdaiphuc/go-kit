package cacheredis

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_meterProviderIsShared(t *testing.T) {
	first, err := meterProvider()
	assert.NoError(t, err)

	second, err := meterProvider()
	assert.NoError(t, err)
	assert.Same(t, first, second, "one Prometheus exporter per process, not per client")
}
