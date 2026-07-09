package publisher

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	publishesTotal     prometheus.Counter
	publishErrorsTotal *prometheus.CounterVec
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
	}
}

func (m *Metrics) IncPublishes() {
	m.publishesTotal.Inc()
}

func (m *Metrics) IncError(reason string) {
	m.publishErrorsTotal.WithLabelValues(reason).Inc()
}
