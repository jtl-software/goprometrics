package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/log"
	"strings"
	"sync"
)

type CounterStore map[string]*prometheus.CounterVec

var lock sync.Mutex

func (s *CounterStore) addCounter(ns string, name string, label ConstLabel) (newCounterCreated bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Error(r)
			err = r.(error)
		}
	}()

	newCounterCreated = false

	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; !ok {
		lock.Lock()
		if _, ok := (*s)[key]; !ok {
			(*s)[key] = promauto.NewCounterVec(
				prometheus.CounterOpts{
					Namespace: ns,
					Name:      name,
				},
				label.Name,
			)
			newCounterCreated = true
			log.Infof("New counter %s_%s with labels %v registered", ns, name, label.Name)
		}
		lock.Unlock()
	}
	return
}

func (s *CounterStore) inc(ns string, name string, label ConstLabel, step float64) {
	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; ok {
		(*s)[key].WithLabelValues(label.Value...).Add(step)
	}
}

func buildKey(ns string, name string, label ConstLabel) string {
	return ns + "_" + name + "__" + strings.Join(label.Name, "_")
}
