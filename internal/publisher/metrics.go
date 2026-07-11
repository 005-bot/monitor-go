package publisher

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	publishesTotal     prometheus.Counter
	publishErrorsTotal *prometheus.CounterVec
	durationSeconds    prometheus.Histogram
}

func NewMetrics() *Metrics {
	return &Metrics{
		publishesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "publisher_publishes_total",
			Help: "Total number of published messages",
		}),
		publishErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "publisher_errors_total",
			Help: "Total number of publish errors",
		}, []string{"reason"}),
		durationSeconds: promauto.NewHistogram(prometheus.HistogramOpts{
			Name: "publisher_duration_seconds",
			Help: "Duration of publish operations",
		}),
	}
}

func (m *Metrics) IncPublishes() {
	m.publishesTotal.Inc()
}

func (m *Metrics) IncError(reason string) {
	m.publishErrorsTotal.WithLabelValues(reason).Inc()
}

func (m *Metrics) ObserveDuration() func() time.Duration {
	timer := prometheus.NewTimer(m.durationSeconds)
	return timer.ObserveDuration
}
