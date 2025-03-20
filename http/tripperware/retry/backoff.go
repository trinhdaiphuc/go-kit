package tripperwareretry

import "time"

// BackoffLinear is very simple: it waits for a fixed period of time between calls.
func BackoffLinear(waitBetween time.Duration) BackoffFunc {
	return func(attempt uint) time.Duration {
		return waitBetween
	}
}
