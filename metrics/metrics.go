package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	MetricBuildInfo = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eck_cluster_eru_limit_build_info",
			Help: "Build information prometheus-ldap-sd ",
		},
		[]string{"version", "git_hash"},
	)
	MetricEruSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "eck_cluster_eru_size_bytes",
			Help: "The number of bytes associated to a single Enterprise Resource Unit (ERU)",
		},
	)
	MetricClusterEruLimit = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			// Namespace: "our_company",
			// Subsystem: "blob_storage",
			Name: "eck_cluster_eru_limit_bytes_total",
			Help: "Total number of ERUs which have been defined for a given cluster",
		},
		[]string{"cluster"},
	)
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func NewResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{w, http.StatusOK}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

// func InitCounter(metric *prometheus.CounterVec, targetGroup string) {
// 	metric.WithLabelValues(targetGroup)
// }
