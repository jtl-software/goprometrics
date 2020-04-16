package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/log"
	"strings"
	"sync"
)

type CounterStore map[string]*prometheus.CounterVec
type SummaryStore map[string]*prometheus.SummaryVec
type HistogramStore map[string]*prometheus.HistogramVec

var counterMutex sync.Mutex
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
