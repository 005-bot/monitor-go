package scraper

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	scrapesTotal      prometheus.Counter
	scrapeErrorsTotal *prometheus.CounterVec
}

func NewMetrics() *Metrics {
	return &Metrics{
		scrapesTotal: promauto.NewCounter(prometheus.CounterOpts{
			Name: "scraper_scrapes_total",
			Help: "Total number of scrapes",
		}),
		scrapeErrorsTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "scraper_errors_total",
			Help: "Total number of scrape errors",
		}, []string{"reason"}),
	}
}

func (m *Metrics) IncTotal() {
	m.scrapesTotal.Inc()
}

func (m *Metrics) IncError(reason string) {
	m.scrapeErrorsTotal.WithLabelValues(reason).Inc()
}
