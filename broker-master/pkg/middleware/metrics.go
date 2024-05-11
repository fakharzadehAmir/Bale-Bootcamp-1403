package middleware

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	ActiveSubscribers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_subscribers",
	})
	MethodCallsError = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "method_count_error",
	}, []string{"method_type"})
	MethodCallsSuccess = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "method_count_success",
	}, []string{"method_type"})
	MethodDuration = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "method_duration",
			Objectives: map[float64]float64{0.50: 0.05, 0.95: 0.01, 0.99: 0.001},
		},
		[]string{"method_type"},
	)
	GarbageCollectorMetric = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "method_gc_broker",
	}, func() float64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return float64(m.GCSys)
	})
)
