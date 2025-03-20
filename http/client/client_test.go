package httpclient

import (
	"context"
	"fmt"
	"net/http"
	"testing"

	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/trinhdaiphuc/go-kit/metrics"
)

func TestNew(t *testing.T) {
	metrics.NewServerMonitor("httpclient")
	http.Handle("/metrics", promhttp.Handler())

	go http.ListenAndServe(":9090", nil)

	cl := NewClient("google")
	for i := 0; i < 1; i++ {
		body, status, err := cl.Get(context.Background(), "https://www.google.com/query?abc=123", nil, nil)
		if err != nil {
			fmt.Println("Call error:", err)
			return
		}

		fmt.Println("Status", status)
		fmt.Println("Body", len(body))
	}

}
