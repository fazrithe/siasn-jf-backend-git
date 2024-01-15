package main

import (
	"github.com/if-itb/siasn-libs-backend/metricutil"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

const MetricNamespace = "siasnJf"
const MetricSubsystem = "generic"

var MetricRespTimeHistogramBucket = []float64{0.025, 0.05, 0.1, 0.25, 0.5, 1, 2.5}

var apiMetrics = &metricutil.DefaultGenericApiPerUrlMetrics{
	RequestTotal: promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace,
		Subsystem: MetricSubsystem,
		Name:      "request_total",
		Help:      "Count the number of service-wide request total with status label success/error",
	}, []string{"path", "status"}),

	RequestInFlight: promauto.NewGaugeVec(prometheus.GaugeOpts{
		Namespace: MetricNamespace,
		Subsystem: MetricSubsystem,
		Name:      "request_in_flight",
		Help:      "Count the number of service-wide in-flight requests.",
	}, []string{"path"}),

	BytesReadTotal: promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace,
		Subsystem: MetricSubsystem,
		Name:      "bytes_read_total",
		Help:      "Count the number of bytes that have been read in service-wide requests.",
	}, []string{"path"}),

	BytesWrittenTotal: promauto.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace,
		Subsystem: MetricSubsystem,
		Name:      "bytes_written_total",
		Help:      "Count the number of bytes that have been written as a response for service-wide requests.",
	}, []string{"path"}),

	RequestDurationSeconds: promauto.NewHistogramVec(prometheus.HistogramOpts{
		Namespace: MetricNamespace,
		Subsystem: MetricSubsystem,
		Name:      "request_duration_seconds",
		Help:      "Measure the duration of processing time of all successful requests service-wide.",
		Buckets:   MetricRespTimeHistogramBucket,
	}, []string{"path"}),
}

var sqlMetrics = &metricutil.DefaultGenericSqlMetrics{
	TxTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: MetricNamespace,
		Subsystem: MetricSubsystem,
		Name:      "tx_total",
		Help:      "The total number of SQL transactions, including those that do not begin with tx.Begin (autocommitted transactions)",
	}, []string{"status"}),

	TxInFlight: prometheus.NewGauge(prometheus.GaugeOpts{
		Namespace: MetricNamespace,
		Subsystem: MetricSubsystem,
		Name:      "tx_in_flight",
		Help:      "The total number of in-flight SQL transactions, including those that do not begin with tx.Begin (autocommitted transactions)",
	}),

	TxDurationSeconds: prometheus.NewHistogram(prometheus.HistogramOpts{
		Namespace: MetricNamespace,
		Subsystem: MetricSubsystem,
		Name:      "tx_duration_seonds",
		Help:      "Measure the duration of processing time of all successful SQL transactions serivce-wide.",
		Buckets:   MetricRespTimeHistogramBucket},
	),
}
