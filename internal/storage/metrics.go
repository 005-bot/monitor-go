package storage

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	opsTotal *prometheus.CounterVec
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
	}
}
