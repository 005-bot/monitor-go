package scheduler

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	cyclesTotal prometheus.Counter
	errorsTotal *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		cyclesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "scheduler_cycles_total",
			Help: "Total number of scheduler cycles",
		}),
		errorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scheduler_errors_total",
			Help: "Total number of scheduler errors",
		}, []string{"reason"}),
	}
}

func (m *Metrics) IncCycles() {
	m.cyclesTotal.Inc()
}

func (m *Metrics) IncError(reason string) {
	m.errorsTotal.WithLabelValues(reason).Inc()
}
