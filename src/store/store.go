package store

import (
	"log/slog"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type (
	Store interface {
		Append(opts MetricOpts)
		Inc(opts MetricOpts, value float64)
		Has(opts MetricOpts) bool
	}
	counterStore struct {
		store map[string]*prometheus.CounterVec
	}
	summaryStore struct {
		store map[string]*prometheus.SummaryVec
	}
	histogramStore struct {
		store map[string]*prometheus.HistogramVec
	}
	gaugeStore struct {
		store map[string]*prometheus.GaugeVec
	}
)

func NewCounterStore() Store {
	return counterStore{
		store: map[string]*prometheus.CounterVec{},
	}
}

func NewSummaryStore() Store {
	return summaryStore{
		store: map[string]*prometheus.SummaryVec{},
	}
}

func NewHistogramStore() Store {
	return histogramStore{
		store: map[string]*prometheus.HistogramVec{},
	}
}

func NewGaugeStore() Store {
	return gaugeStore{
		store: map[string]*prometheus.GaugeVec{},
	}
}

// Counter
func (s counterStore) Has(opts MetricOpts) bool {
	_, has := s.store[opts.Key()]
	return has
}

func (s counterStore) Append(opts MetricOpts) {
	s.store[opts.Key()] = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: opts.Ns,
			Name:      opts.Name,
			Help:      opts.Help,
		},
		opts.Label.Name,
	)
	slog.Info("New counter registered", "ns", opts.Ns, "name", opts.Name, "labels", opts.Label.Name)
}

func (s counterStore) Inc(opts MetricOpts, value float64) {
	if s.Has(opts) {
		s.store[opts.Key()].WithLabelValues(opts.Label.Value...).Add(value)
	}
}

// Summary
func (s summaryStore) Has(opts MetricOpts) bool {
	_, has := s.store[opts.Key()]
	return has
}

func (s summaryStore) Append(opts MetricOpts) {
	s.store[opts.Key()] = promauto.NewSummaryVec(
		prometheus.SummaryOpts{
			Namespace:  opts.Ns,
			Name:       opts.Name,
			Help:       opts.Help,
			Objectives: opts.SummaryObjectives,
		},
		opts.Label.Name,
	)
	slog.Info("New summary registered", "ns", opts.Ns, "name", opts.Name, "labels", opts.Label.Name, "objectives", opts.SummaryObjectives)
}

func (s summaryStore) Inc(opts MetricOpts, value float64) {
	if s.Has(opts) {
		s.store[opts.Key()].WithLabelValues(opts.Label.Value...).Observe(value)
	}
}

// Histogram
func (s histogramStore) Has(opts MetricOpts) bool {
	_, has := s.store[opts.Key()]
	return has
}

func (s histogramStore) Append(opts MetricOpts) {
	s.store[opts.Key()] = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Namespace: opts.Ns,
			Name:      opts.Name,
			Help:      opts.Help,
			Buckets:   opts.HistogramBuckets,
		},
		opts.Label.Name,
	)
	slog.Info("New histogram registered", "ns", opts.Ns, "name", opts.Name, "labels", opts.Label.Name, "buckets", opts.HistogramBuckets)
}

func (s histogramStore) Inc(opts MetricOpts, value float64) {
	if s.Has(opts) {
		s.store[opts.Key()].WithLabelValues(opts.Label.Value...).Observe(value)
	}
}

// Gauge
func (g gaugeStore) Append(opts MetricOpts) {
	g.store[opts.Key()] = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Namespace: opts.Ns,
			Name:      opts.Name,
			Help:      opts.Help,
		},
		opts.Label.Name,
	)
	slog.Info("New gauge registered", "ns", opts.Ns, "name", opts.Name, "labels", opts.Label.Name)
}

func (g gaugeStore) Inc(opts MetricOpts, value float64) {
	if g.Has(opts) {
		gauge := g.store[opts.Key()].WithLabelValues(opts.Label.Value...)

		if opts.SetGaugeToValue == true {
			gauge.Set(value)
		} else {
			gauge.Add(value)
		}
	}
}

func (g gaugeStore) Has(opts MetricOpts) bool {
	_, has := g.store[opts.Key()]
	return has
}
