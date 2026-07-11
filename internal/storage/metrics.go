package storage

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	opsTotal        *prometheus.CounterVec
	durationSeconds *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		opsTotal: promauto.NewCounterVec(
			prometheus.CounterOpts{
				Name: "storage_ops_total",
				Help: "Total number of storage operations",
			},
			[]string{"op"},
		),
		durationSeconds: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "storage_duration_seconds",
			Help:    "Duration of storage operations",
			Buckets: []float64{0.0005, 0.001, 0.0025, 0.005, 0.01, 0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5, 5},
		}, []string{"op"}),
	}
}

func (m *Metrics) ObserveDuration(op string, seconds float64) {
	m.durationSeconds.WithLabelValues(op).Observe(seconds)
}
