package report

import "time"

type options struct {
	prometheusQueryTimeout time.Duration
	progressReporter       progressReporter
}

// Option represents a report option.
type Option interface {
	set(*options)
}

func buildOptions(os []Option) options {
	var build options
	for _, o := range os {
		o.set(&build)
	}
	return build
}

// WithPrometheusQueryTimeout allows setting a timout when querying prometheus.
func WithPrometheusQueryTimeout(tm time.Duration) Option {
	return prometheusQueryTimeout(tm)
}

type prometheusQueryTimeout time.Duration

func (t prometheusQueryTimeout) set(o *options) {
	o.prometheusQueryTimeout = time.Duration(t)
}

// Progress represent the progress when generating multiple reports.
type Progress struct {
	Timestamp time.Time
	Count     int
}

// WithProgressReporter allows setting a callback function.
// The callback receives a progress report when generating multiple reports.
func WithProgressReporter(r func(Progress)) Option {
	return progressReporter(r)
}

type progressReporter func(Progress)

func (t progressReporter) set(o *options) {
	o.progressReporter = t
}
