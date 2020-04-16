package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/log"
	"strings"
	"sync"
)

type PrometheusMetricOpts struct {
	ns string
	name string
	label ConstLabel
	help string
	histogramBuckets []float64
	summaryObjectives map[float64]float64
}

func (opts PrometheusMetricOpts) buildStoreKey() string {
	return opts.ns + "_" + opts.name + "__" + strings.Join(opts.label.Name, "_")
}

type PrometheusMetricAdapter interface {
	create(opts PrometheusMetricOpts, creator func(opts PrometheusMetricOpts, s *CounterStore)) (newCreated bool, err error)
	invoke(opts PrometheusMetricOpts, value float64)
	has(opts PrometheusMetricOpts) bool
	append(opts PrometheusMetricOpts)
}

type NewCounterStore struct {
	store map[string]*prometheus.CounterVec
}

func test()  {

	store := NewCounterStore{
		store:  map[string]*prometheus.CounterVec{},
	}

	opts := PrometheusMetricOpts{
		ns:                "foo",
		name:              "name",
		label:             ConstLabel{},
		help:              "help",
		histogramBuckets:  nil,
		summaryObjectives: nil,
	}
	createCounter := func(opts PrometheusMetricOpts, s *PrometheusMetricAdapter) {
		s.append(opts)
		(*s)[opts.buildStoreKey()] = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: opts.ns,
				Name:      opts.name,
				Help:      opts.help,
			},
			opts.label.Name,
		)
		log.Infof("A new counter %s_%s with labels %v registered", opts.ns, opts.name, opts.label.Name)
	}
	store.create(opts, createCounter)
}

func (s *NewCounterStore) append(opts PrometheusMetricOpts)  {
	s.store[opts.buildStoreKey()] = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: opts.ns,
			Name:      opts.name,
			Help:      opts.help,
		},
		opts.label.Name,
	)
}

func (s *NewCounterStore) has(opts PrometheusMetricOpts) bool {
	_, has := s.store[opts.buildStoreKey()]
	return has
}

func (s *NewCounterStore) create(opts PrometheusMetricOpts, creator func(opts PrometheusMetricOpts, s *PrometheusMetricAdapter)) (newCreated bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			err = r.(error)
		}
	}()

	newCreated = false
	if !s.has(opts) {
		counterMutex.Lock()
		defer counterMutex.Unlock()
		if !s.has(opts) {
			creator(opts, s)
			newCreated = true
		}
	}
	return
}

func (s *NewCounterStore) invoke(opts PrometheusMetricOpts, value float64) {

}


type CounterStore map[string]*prometheus.CounterVec
type SummaryStore map[string]*prometheus.SummaryVec
type HistogramStore map[string]*prometheus.HistogramVec

var counterMutex sync.Mutex
// old stuff
var summaryMutex sync.Mutex
var histogramMutex sync.Mutex

func (s *CounterStore) addCounter(ns string, name string, label ConstLabel, help string) (newCounterCreated bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			err = r.(error)
		}
	}()

	newCounterCreated = false

	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; !ok {
		counterMutex.Lock()
		if _, ok := (*s)[key]; !ok {
			(*s)[key] = promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: ns,
					Name:      name,
					Help:      help,
				},
				label.Name,
			)
			newCounterCreated = true
			log.Infof("A new counter %s_%s with labels %v registered", ns, name, label.Name)
		}
		counterMutex.Unlock()
	}
	return
}

func (s *CounterStore) inc(ns string, name string, label ConstLabel, step float64) {
	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; ok {
		(*s)[key].WithLabelValues(label.Value...).Add(step)
	}
}

func (s *SummaryStore) addSummary(ns string, name string, label ConstLabel, objectives map[float64]float64, help string) (newSummaryCreated bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			err = r.(error)
		}
	}()

	newSummaryCreated = false

	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; !ok {
		summaryMutex.Lock()
		if _, ok := (*s)[key]; !ok {
			(*s)[key] = promauto.NewSummaryVec(
				prometheus.SummaryOpts{
					Namespace:  ns,
					Name:       name,
					Help:       help,
					Objectives: objectives,
				},
				label.Name,
			)
			newSummaryCreated = true
			log.Infof("A new summary %s_%s with labels %v and objectives %v registered", ns, name, label.Name, objectives)
		}
		summaryMutex.Unlock()
	}
	return

}

func (s *SummaryStore) observe(ns string, name string, label ConstLabel, observation float64) {
	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; ok {
		(*s)[key].WithLabelValues(label.Value...).Observe(observation)
	}
}

func (s *HistogramStore) addHistogram(ns string, name string, label ConstLabel, buckets []float64, help string) (newHistogramCreated bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			err = r.(error)
		}
	}()

	newHistogramCreated = false

	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; !ok {
		histogramMutex.Lock()
		if _, ok := (*s)[key]; !ok {
			(*s)[key] = promauto.NewHistogramVec(
				prometheus.HistogramOpts{
					Namespace: ns,
					Name:      name,
					Help:      help,
					Buckets:   buckets,
				},
				label.Name,
			)
			newHistogramCreated = true
			log.Infof("A new histogram %s_%s with labels %v and buckets %v registered", ns, name, label.Name, buckets)
		}
		histogramMutex.Unlock()
	}
	return

}

func (s *HistogramStore) observe(ns string, name string, label ConstLabel, observation float64) {
	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; ok {
		(*s)[key].WithLabelValues(label.Value...).Observe(observation)
	}
}

func buildKey(ns string, name string, label ConstLabel) string {
	return ns + "_" + name + "__" + strings.Join(label.Name, "_")
}
