package metricutil

import (
	"github.com/prometheus/client_golang/prometheus"
	"time"
)

type DurationObserver struct {
	start            time.Time
	flightGauge      prometheus.Gauge
	responseDuration prometheus.Observer
}

// Deprecated: just create these observers manually.
// NewDurationObserver is constructor of DurationObserver.
func NewDurationObserver(fg prometheus.Gauge, rd prometheus.Observer) *DurationObserver {
	observer := &DurationObserver{
		flightGauge:      fg,
		responseDuration: rd,
	}

	return observer
}

// This method is used to initialize flight gauge and response duration start time.
func (observer *DurationObserver) Start() {
	observer.start = time.Now()
	observer.flightGauge.Inc()
}

// This method is used to end an observation for DurationObserver as well as decreasing in flight metrics.
func (observer *DurationObserver) End() {
	observer.flightGauge.Dec()
	observer.responseDuration.Observe(time.Since(observer.start).Seconds())
}

// StartObserver is used to initialize a new DurationObserver and start it.
func StartObserver(fg prometheus.Gauge, rd prometheus.Observer) *DurationObserver {
	observer := NewDurationObserver(fg, rd)
	observer.Start()
	return observer
}
