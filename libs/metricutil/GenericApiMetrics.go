package metricutil

import (
	"github.com/miolini/datacounter"
	"github.com/prometheus/client_golang/prometheus"
	"io"
	"net/http"
	"time"
)

// GenericApiMetrics define a type containing five types of generic API measurement metrics: request total (with status),
// bytes read total, bytes written total, processing duration, and in-flight metric.
// Implementers can create a struct with these three metrics and implement this interface. Services can then use this
// interface anywhere.
type GenericApiMetrics interface {
	IncRequestTotalSuccess()
	IncRequestTotalError()
	AddBytesReadTotal(bytes int)
	AddBytesWrittenTotal(bytes int)
	IncInFlight()
	DecInFlight()
	// ObserveRequestDuration only measures successful requests.
	ObserveRequestDuration(duration float64)
}

// GenericApiMetricsPerUrl define a type containing five types of generic API measurement metrics: request total (with status),
// bytes read total, bytes written total, processing duration, and in-flight metric.
// Implementers can create a struct with these three metrics and implement this interface. Services can then use this
// interface anywhere.
//
// The difference between this and GenericApiMetrics is that this interface forces the implementer to store metric
// per API path. It is expected, for example, that `/api/v1/login` path will have different counter with `/api/v1/submit`.
// GenericApiMetrics aggregates requests from all paths into a single metric.
type GenericApiMetricsPerUrl interface {
	// InitializeLabel initialize metrics with label named path.
	// This is done so that the metrics will show in /metrics page even though no metrics have been added/decremented.
	// You do not need to call this and initialize all paths that you know. This is done usually for cosmetic purposes.
	InitializeLabel(path string)
	// IncRequestTotalSuccess increments total success request per URL path. URL path must not contain scheme, query parameters, and so on.
	// The path will be used as vector label, so `/api/v1/login` path will have different counter with `/api/v1/submit`
	// for example.
	IncRequestTotalSuccess(path string)
	// IncRequestTotalError increments total error request per URL path. URL path must not contain scheme, query parameters, and so on.
	// The path will be used as vector label, so `/api/v1/login` path will have different counter with `/api/v1/submit`
	// for example.
	IncRequestTotalError(path string)
	AddBytesReadTotal(path string, bytes int)
	AddBytesWrittenTotal(path string, bytes int)
	IncInFlight(path string)
	DecInFlight(path string)
	// ObserveRequestDuration only measures successful requests.
	ObserveRequestDuration(path string, duration float64)
}

// DefaultGenericApiMetrics is a default structure that implements GenericApiMetrics.
// GenericApiMetrics is used to measure service-wide performance (all API endpoints are measured as a single metric).
type DefaultGenericApiMetrics struct {
	RequestTotal           *prometheus.CounterVec
	RequestInFlight        prometheus.Gauge
	BytesReadTotal         prometheus.Counter
	BytesWrittenTotal      prometheus.Counter
	RequestDurationSeconds prometheus.Observer
}

func (gam *DefaultGenericApiMetrics) IncRequestTotalSuccess() {
	gam.RequestTotal.WithLabelValues("success").Inc()
}

func (gam *DefaultGenericApiMetrics) IncRequestTotalError() {
	gam.RequestTotal.WithLabelValues("error").Inc()
}

func (gam *DefaultGenericApiMetrics) AddBytesReadTotal(bytes int) {
	gam.BytesReadTotal.Add(float64(bytes))
}

func (gam *DefaultGenericApiMetrics) AddBytesWrittenTotal(bytes int) {
	gam.BytesWrittenTotal.Add(float64(bytes))
}

func (gam *DefaultGenericApiMetrics) IncInFlight() {
	gam.RequestInFlight.Inc()
}

func (gam *DefaultGenericApiMetrics) DecInFlight() {
	gam.RequestInFlight.Dec()
}

func (gam *DefaultGenericApiMetrics) ObserveRequestDuration(duration float64) {
	gam.RequestDurationSeconds.Observe(duration)
}

// DefaultGenericApiPerUrlMetrics is a default structure that implements GenericApiMetricsPerUrl.
// GenericApiMetricsPerUrl is used to measure service-wide performance (all API endpoints are measured as a single metric).
// It stores each metric per URL path vector label, in other words each API path will have different metrics.
//
// Be careful as not all path vector labels are initialized when the service is run first. This is because there needs
// to be at least one call to each method this struct has, per path, to initialize its label value. When a path label has
// not been initialized, it will not appear in /metrics page. See example below.
//
// For `/api/v1/login` for example, its metrics are stored as vector label `/api/v1/login`. Now the /metrics page,
// when the service is run, will not show any metrics with label `/api/v1/login`, not until a call to /api/v1/login
// has been received, which will initialized the metric and so it will appear on /metrics page. It is possible
// to initialize all known label values though if the paths have been known beforehand by using InitializeLabel.
type DefaultGenericApiPerUrlMetrics struct {
	RequestTotal           *prometheus.CounterVec
	RequestInFlight        *prometheus.GaugeVec
	BytesReadTotal         *prometheus.CounterVec
	BytesWrittenTotal      *prometheus.CounterVec
	RequestDurationSeconds prometheus.ObserverVec
}

// InitializeLabel initializes all metrics to have label path.
func (gam *DefaultGenericApiPerUrlMetrics) InitializeLabel(path string) {
	gam.RequestTotal.WithLabelValues(path, "success")
	gam.RequestTotal.WithLabelValues(path, "error")
	gam.BytesReadTotal.WithLabelValues(path)
	gam.BytesWrittenTotal.WithLabelValues(path)
	gam.RequestInFlight.WithLabelValues(path)
	gam.RequestDurationSeconds.WithLabelValues(path)
}

func (gam *DefaultGenericApiPerUrlMetrics) IncRequestTotalSuccess(path string) {
	gam.RequestTotal.WithLabelValues(path, "success").Inc()
}

func (gam *DefaultGenericApiPerUrlMetrics) IncRequestTotalError(path string) {
	gam.RequestTotal.WithLabelValues(path, "error").Inc()
}

func (gam *DefaultGenericApiPerUrlMetrics) AddBytesReadTotal(path string, bytes int) {
	gam.BytesReadTotal.WithLabelValues(path).Add(float64(bytes))
}

func (gam *DefaultGenericApiPerUrlMetrics) AddBytesWrittenTotal(path string, bytes int) {
	gam.BytesWrittenTotal.WithLabelValues(path).Add(float64(bytes))
}

func (gam *DefaultGenericApiPerUrlMetrics) IncInFlight(path string) {
	gam.RequestInFlight.WithLabelValues(path).Inc()
}

func (gam *DefaultGenericApiPerUrlMetrics) DecInFlight(path string) {
	gam.RequestInFlight.WithLabelValues(path).Dec()
}

func (gam *DefaultGenericApiPerUrlMetrics) ObserveRequestDuration(path string, duration float64) {
	gam.RequestDurationSeconds.WithLabelValues(path).Observe(duration)
}

// GenericApiResponseWriter wraps a http.ResponseWriter and automatically adds request total metrics and bytes written
// metrics. The success request total is added when 2xx codes is received, while error is added for 4xx and 5xx codes.
type GenericApiResponseWriter struct {
	writer            http.ResponseWriter
	genericApiMetrics GenericApiMetrics
	startTime         time.Time
	lastStatusCode    int
}

func NewGenericApiResponseWriter(writer http.ResponseWriter, genericApiMetrics GenericApiMetrics) *GenericApiResponseWriter {
	return &GenericApiResponseWriter{writer: writer, genericApiMetrics: genericApiMetrics, lastStatusCode: http.StatusOK}
}

// StartObserving start the observation of duration and in-flight metrics.
func (g *GenericApiResponseWriter) StartObserving() {
	g.startTime = time.Now()
	g.genericApiMetrics.IncInFlight()
}

// EndObserving stops the observation of duration and in-flight metrics, effectively decrementing the in-flight metrics,
// and registering the duration metrics.
func (g *GenericApiResponseWriter) EndObserving() {
	if g.lastStatusCode >= 200 && g.lastStatusCode < 300 {
		g.genericApiMetrics.ObserveRequestDuration(time.Since(g.startTime).Seconds())
	}
	g.genericApiMetrics.DecInFlight()
}

func (g *GenericApiResponseWriter) Header() http.Header {
	return g.writer.Header()
}

func (g *GenericApiResponseWriter) Write(bytes []byte) (n int, err error) {
	n, err = g.writer.Write(bytes)
	g.genericApiMetrics.AddBytesWrittenTotal(n)
	return
}

func (g *GenericApiResponseWriter) WriteHeader(statusCode int) {
	g.lastStatusCode = statusCode
	g.writer.WriteHeader(statusCode)
	if statusCode >= 200 && statusCode < 300 {
		g.genericApiMetrics.IncRequestTotalSuccess()
	} else if statusCode >= 400 {
		g.genericApiMetrics.IncRequestTotalError()
	}
}

func NewGenericApiPerUrlResponseWriter(writer http.ResponseWriter, path string, genericApiMetrics GenericApiMetricsPerUrl) *GenericApiPerUrlResponseWriter {
	return &GenericApiPerUrlResponseWriter{writer: writer, path: path, genericApiMetrics: genericApiMetrics, lastStatusCode: http.StatusOK}
}

// GenericApiPerUrlResponseWriter wraps a http.ResponseWriter and automatically adds request total metrics and bytes written
// metrics. The success request total is added when 2xx codes is received, while error is added for 4xx and 5xx codes.
//
// This ResponseWriter is created per handler. It will read different path for each handler. That is why the constructor
// requires path to be supplied. It is usually fetched from (*http.Request).URL.Path. The writer will store the path
// until the handler is completed.
type GenericApiPerUrlResponseWriter struct {
	writer            http.ResponseWriter
	path              string
	genericApiMetrics GenericApiMetricsPerUrl
	startTime         time.Time
	lastStatusCode    int
	lastTotalWritten  int
}

// StartObserving start the observation of duration and in-flight metrics.
func (g *GenericApiPerUrlResponseWriter) StartObserving() {
	g.startTime = time.Now()
	g.genericApiMetrics.IncInFlight(g.path)
}

// EndObserving stops the observation of duration and in-flight metrics, effectively decrementing the in-flight metrics,
// and registering the duration metrics.
//
// It will also flush total bytes written and add total bytes written metrics.
func (g *GenericApiPerUrlResponseWriter) EndObserving() {
	if g.lastStatusCode >= 200 && g.lastStatusCode < 300 {
		g.genericApiMetrics.ObserveRequestDuration(g.path, time.Since(g.startTime).Seconds())
	}
	g.genericApiMetrics.DecInFlight(g.path)
	g.genericApiMetrics.AddBytesWrittenTotal(g.path, g.lastTotalWritten)
	g.lastTotalWritten = 0
}

func (g *GenericApiPerUrlResponseWriter) Header() http.Header {
	return g.writer.Header()
}

func (g *GenericApiPerUrlResponseWriter) Write(bytes []byte) (n int, err error) {
	n, err = g.writer.Write(bytes)
	// Instead of adding metrics directly, store total bytes first.
	g.lastTotalWritten += n
	return
}

func (g *GenericApiPerUrlResponseWriter) WriteHeader(statusCode int) {
	g.lastStatusCode = statusCode
	g.writer.WriteHeader(statusCode)
	if statusCode >= 200 && statusCode < 300 {
		g.genericApiMetrics.IncRequestTotalSuccess(g.path)
	} else if statusCode >= 400 {
		g.genericApiMetrics.IncRequestTotalError(g.path)
	}
}

// RequestBodyCounter is a wrapper around request.Body.
// This will count any read action against the body.
// See also GenericApiMetricsWrapper and GenericApiMetricsPerUrlWrapper.
type RequestBodyCounter struct {
	*datacounter.ReaderCounter
	io.Closer
}

func NewRequestBodyCounter(body io.ReadCloser) *RequestBodyCounter {
	return &RequestBodyCounter{datacounter.NewReaderCounter(body), body}
}

// GenericApiMetricsWrapper is a HTTP wrapper to wrap a handler and replace its writer with GenericApiResponseWriter to automatically measure
// generic metrics.
func GenericApiMetricsWrapper(handler http.Handler, metrics GenericApiMetrics) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := NewGenericApiResponseWriter(w, metrics)
		writer.StartObserving()
		counter := NewRequestBodyCounter(r.Body)
		r.Body = counter
		defer func() {
			writer.EndObserving()
			metrics.AddBytesReadTotal(int(counter.Count()))
		}()
		handler.ServeHTTP(writer, r)
	})
}

// GenericApiMetricsPerUrlWrapper is an HTTP wrapper to wrap a handler and replace its writer with
// GenericApiPerUrlResponseWriter to automatically measure generic metrics.
//
// GenericApiPerUrlResponseWriter does NOT interact with BytesRead metrics. This is because, to count
// bytes read you need to read the request payload first, which may not be done, or may not be what you want. The wrapper
// simply does not know what you read and do not read.
func GenericApiMetricsPerUrlWrapper(handler http.Handler, metrics GenericApiMetricsPerUrl) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		writer := NewGenericApiPerUrlResponseWriter(w, r.URL.Path, metrics)
		writer.StartObserving()
		counter := NewRequestBodyCounter(r.Body)
		r.Body = counter
		defer func() {
			writer.EndObserving()
			metrics.AddBytesReadTotal(r.URL.Path, int(counter.Count()))
		}()
		handler.ServeHTTP(writer, r)
	})
}
