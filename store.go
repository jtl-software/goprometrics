package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/log"
	"strings"
	"sync"
)

var metricCreateMutex sync.Mutex

type MetricStore struct {
	CounterStore   map[string]*prometheus.CounterVec
	SummaryStore   map[string]*prometheus.SummaryVec
	HistogramStore map[string]*prometheus.HistogramVec
}

type PrometheusMetricOpts struct {
	ns                string
	name              string
	label             ConstLabel
	help              string
	histogramBuckets  []float64
	summaryObjectives map[float64]float64
}

func (opts PrometheusMetricOpts) BuildStoreKey() string {
	return opts.ns + "_" + opts.name + "__" + strings.Join(opts.label.Name, "_")
}

func NewMetricStore() MetricStore {
	return MetricStore{
		CounterStore:   map[string]*prometheus.CounterVec{},
		SummaryStore:   map[string]*prometheus.SummaryVec{},
		HistogramStore: map[string]*prometheus.HistogramVec{},
	}
}

func (s *MetricStore) has(opts PrometheusMetricOpts) bool {
	_, hasCounter := s.CounterStore[opts.BuildStoreKey()]
	_, hasSummary := s.SummaryStore[opts.BuildStoreKey()]
	_, hasHistogram := s.HistogramStore[opts.BuildStoreKey()]
	return hasCounter || hasSummary || hasHistogram
}

func (s *MetricStore) append(opts PrometheusMetricOpts, creator func(opts PrometheusMetricOpts, s *MetricStore)) (newCreated bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			err = r.(error)
			log.Error(r)
		}
	}()

	newCreated = false
	if !s.has(opts) {
		metricCreateMutex.Lock()
		defer metricCreateMutex.Unlock()
		if !s.has(opts) {
			creator(opts, s)
			newCreated = true
		}
	}
	return
}

func CreateCounterMetricHandler() func(opts PrometheusMetricOpts, s *MetricStore) {
	return func(opts PrometheusMetricOpts, s *MetricStore) {
		s.CounterStore[opts.BuildStoreKey()] = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: opts.ns,
				Name:      opts.name,
				Help:      opts.help,
			},
			opts.label.Name,
		)
		log.Infof("A new counter %s_%s with labels %v registered", opts.ns, opts.name, opts.label.Name)
	}
}

func CreateSummaryMetricHandler() func(opts PrometheusMetricOpts, s *MetricStore) {
	return func(opts PrometheusMetricOpts, s *MetricStore) {
		s.SummaryStore[opts.BuildStoreKey()] = promauto.NewSummaryVec(
			prometheus.SummaryOpts{
				Namespace:  opts.ns,
				Name:       opts.name,
				Help:       opts.help,
				Objectives: opts.summaryObjectives,
			},
			opts.label.Name,
		)
		log.Infof("A new summary %s_%s with labels %v and objectives %v registered", opts.ns, opts.name, opts.label.Name, opts.summaryObjectives)
	}
}

func CreateHistogramMetricHandler() func(opts PrometheusMetricOpts, s *MetricStore) {
	return func(opts PrometheusMetricOpts, s *MetricStore) {
		s.HistogramStore[opts.BuildStoreKey()] = promauto.NewHistogramVec(
			prometheus.HistogramOpts{
				Namespace: opts.ns,
				Name:      opts.name,
				Help:      opts.help,
				Buckets:   opts.histogramBuckets,
			},
			opts.label.Name,
		)
		log.Infof("A new histogram %s_%s with labels %v and buckets %v registered", opts.ns, opts.name, opts.label.Name, opts.histogramBuckets)
	}
}

type ConstLabel struct {
	Name  []string
	Value []string
}

func IncCounterHandler() func(s *MetricStore, opts PrometheusMetricOpts, value float64) {
	return func(s *MetricStore, opts PrometheusMetricOpts, value float64) {
		s.CounterStore[opts.BuildStoreKey()].WithLabelValues(opts.label.Value...).Add(value)
	}
}

func ObserveSummaryHandler() func(s *MetricStore, opts PrometheusMetricOpts, value float64) {
	return func(s *MetricStore, opts PrometheusMetricOpts, value float64) {
		s.SummaryStore[opts.BuildStoreKey()].WithLabelValues(opts.label.Value...).Observe(value)
	}
}

func ObserveHistogramHandler() func(s *MetricStore, opts PrometheusMetricOpts, value float64) {
	return func(s *MetricStore, opts PrometheusMetricOpts, value float64) {
		s.HistogramStore[opts.BuildStoreKey()].WithLabelValues(opts.label.Value...).Observe(value)
	}
}
