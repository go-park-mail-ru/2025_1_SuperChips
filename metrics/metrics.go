package metrics

import (
	"context"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"google.golang.org/grpc"
)

type MetricsService struct {
	hits      *prometheus.CounterVec
	errors    *prometheus.CounterVec
	durations *prometheus.HistogramVec
}

func NewMetricsService() *MetricsService {
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

func (m *MetricsService) ServerMetricsInterceptor(
	ctx context.Context,
	req interface{},
	info *grpc.UnaryServerInfo,
	handler grpc.UnaryHandler,
) (interface{}, error) {
	serviceName := extractServiceName(info.FullMethod)

	start := time.Now()

	h, err := handler(ctx, req)

	m.AddDurationToHistogram(serviceName, info.FullMethod, time.Since(start))
	m.IncreaseHits(serviceName, info.FullMethod, strconv.Itoa(200))
	if err != nil {
		m.IncreaseErr(serviceName, info.FullMethod, serviceName + " : " + err.Error())
	}

	return h, err
}

func extractServiceName(fullMethod string) string {
	parts := strings.Split(fullMethod, "/")
	if len(parts) >= 2 {
		if len(parts[1]) >= 2 {
			return strings.Split(parts[1], ".")[0]
		}
	}
	return "unknown"
}

