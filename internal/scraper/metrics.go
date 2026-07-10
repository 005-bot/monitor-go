package scraper

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type Metrics struct {
	scrapesTotal      prometheus.Counter
	scrapeErrorsTotal *prometheus.CounterVec
	durationSeconds   prometheus.Histogram
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
		durationSeconds: promauto.NewHistogram(prometheus.HistogramOpts{
			Name: "scraper_duration_seconds",
			Help: "Duration of scrape operations",
		}),
	}
}

func (m *Metrics) IncTotal() {
	m.scrapesTotal.Inc()
}

func (m *Metrics) IncError(reason string) {
	m.scrapeErrorsTotal.WithLabelValues(reason).Inc()
}

func (m *Metrics) ObserveDuration(seconds float64) {
	m.durationSeconds.Observe(seconds)
}
