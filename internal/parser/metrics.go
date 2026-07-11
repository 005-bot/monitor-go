package parser

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	operationsTotal *prometheus.CounterVec
	durationSeconds *prometheus.HistogramVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		operationsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "parser_operations_total",
			Help: "Total number of parser operations",
		}, []string{"op"}),
		durationSeconds: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name: "parser_duration_seconds",
			Help: "Duration of parser operations",
		}, []string{"op"}),
	}
}

func (m *Metrics) IncOperations(op string) {
	m.operationsTotal.WithLabelValues(op).Inc()
}

func (m *Metrics) ObserveDuration(op string) func() time.Duration {
	start := prometheus.NewTimer(m.durationSeconds.WithLabelValues(op))
	return start.ObserveDuration
}
