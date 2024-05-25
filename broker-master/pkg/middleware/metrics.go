package middleware

import (
	"runtime"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/load"
)

var (
	ActiveSubscribers = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_subscribers",
	})

	MethodCount = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name: "method_count",
		},
		[]string{"method", "status"},
	)
	MethodDuration = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Name:       "method_duration",
			Objectives: map[float64]float64{0.50: 0.05, 0.95: 0.01, 0.99: 0.001},
		},
		[]string{"method_type"},
	)

	MemoryUsage = promauto.NewGaugeFunc(prometheus.GaugeOpts{
		Name: "memory_usage_bytes",
	}, func() float64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return float64(m.Alloc)
	})

	NumGC = promauto.NewCounterFunc(prometheus.CounterOpts{
		Name: "total_gc",
	}, func() float64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return float64(m.NumGC)
	})

	SystemCPUUtilization = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "cpu_utilization_percent",
	}, []string{"cpu"})

	CPULoad1m = promauto.NewGauge(prometheus.GaugeOpts{Name: "cpu_load"})
)

func EvaluateEnvMetrics() {
	ticker := time.NewTicker(500 * time.Millisecond)
	for {
		select {
		case <-ticker.C:
			loadAvg, _ := load.Avg()
			CPULoad1m.Set(loadAvg.Load1)

			cpuPercent, _ := cpu.Percent(time.Second, true)
			for idx, cp := range cpuPercent {
				SystemCPUUtilization.With(prometheus.Labels{"cpu": strconv.Itoa(idx)}).Set(cp)
			}
		}
	}
}
