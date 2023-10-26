package internal

import "time"

type Config struct {
	Interval           time.Duration
	PrometheusEndpoint string
	ExcludedMetrics    []string
}
