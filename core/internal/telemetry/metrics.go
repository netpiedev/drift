package telemetry

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type Metrics struct {
	migrationDuration *prometheus.HistogramVec
	migrationTotal    *prometheus.CounterVec
	once              sync.Once
}

func NewMetrics() *Metrics {
	m := &Metrics{
		migrationDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "drift",
			Name:      "migration_duration_seconds",
			Help:      "Duration of migration execution.",
			Buckets:   prometheus.DefBuckets,
		}, []string{"name", "direction", "success"}),
		migrationTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Namespace: "drift",
			Name:      "migration_total",
			Help:      "Count of migrations executed.",
		}, []string{"direction", "success"}),
	}
	return m
}

func (m *Metrics) Register() {
	m.once.Do(func() {
		prometheus.MustRegister(m.migrationDuration, m.migrationTotal)
	})
}

func (m *Metrics) ObserveMigration(name string, direction string, success bool, duration time.Duration) {
	s := boolString(success)
	m.migrationDuration.WithLabelValues(name, direction, s).Observe(duration.Seconds())
	m.migrationTotal.WithLabelValues(direction, s).Inc()
}

func StartPrometheusServer(addr string) error {
	if addr == "" {
		return nil
	}
	h := http.NewServeMux()
	h.Handle("/metrics", promhttp.Handler())
	go func() {
		_ = http.ListenAndServe(addr, h)
	}()
	return nil
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func DescribeLockWait(waitMS int64) string {
	return fmt.Sprintf("%dms", waitMS)
}
