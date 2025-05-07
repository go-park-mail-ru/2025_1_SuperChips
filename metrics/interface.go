package metrics

import (
	"time"
)

type MetricsServicer interface {
	IncreaseHits(method string, path string, statusCode string)
	IncreaseErr(method string, path string, service string)
	AddDurationToHistogram(method string, service string, duration time.Duration)

	RegisterMetrics()
}