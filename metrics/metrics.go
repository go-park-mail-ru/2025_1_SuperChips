package metrics

import (
	"time"

	"github.com/go-park-mail-ru/2025_1_SuperChips/configs"
	"github.com/prometheus/client_golang/prometheus"
)

type MetricsService struct {
	hits      *prometheus.CounterVec
	errors    *prometheus.CounterVec
	durations *prometheus.HistogramVec
}

func NewMetricsService(cfg configs.Config) *MetricsService {
	return &MetricsService{
		hits: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_method_hits_total",
			Help: "Total number of http method calls across all services",
		}, []string{"method", "path", "status_code"}),

		errors: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: "http_method_errors_total",
			Help: "Total number of http method errors across all services",
		}, []string{"method", "path", "description"}),

		durations: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "http_method_duration_seconds",
			Help:    "Histogram of http method call durations across services",
			Buckets: prometheus.DefBuckets,
		}, []string{"method", "service"}),
	}
}

func (m *MetricsService) RegisterMetrics() {
	prometheus.MustRegister(m.hits)
	prometheus.MustRegister(m.errors)
	prometheus.MustRegister(m.durations)
}

func (m *MetricsService) IncreaseHits(method string, path string, statusCode string) {
	m.hits.WithLabelValues(method, path, statusCode).Inc()
}

func (m *MetricsService) IncreaseErr(method string, path string, description string) {
	m.errors.WithLabelValues(method, path, description).Inc()
}

func (m *MetricsService) AddDurationToHistogram(method, service string, duration time.Duration) {
	m.durations.WithLabelValues(method, service).Observe(duration.Seconds())
}
