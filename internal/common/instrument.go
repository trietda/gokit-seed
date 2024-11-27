package common

import (
	"time"

	"github.com/go-kit/kit/metrics"
)

type Metrics struct {
	Count   metrics.Counter
	Latency metrics.Histogram
}

func NewMetrics(count metrics.Counter, latency metrics.Histogram) Metrics {
	return Metrics{
		Count:   count,
		Latency: latency,
	}
}

func (m Metrics) Collect(completeTime time.Time, labelValues ...string) {
	m.Count.With(labelValues...).Add(1)
	m.Latency.With(labelValues...).Observe(time.Since(completeTime).Seconds())
}
