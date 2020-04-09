package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/common/log"
	"strings"
)

type CounterStore map[string]*prometheus.CounterVec

func (s *CounterStore) addCounter(ns string, name string, label ConstLabel) bool {
	newCounterCreated := false

	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; !ok {
		log.Infof("New counter %s_%s registered", ns, name)

		newCounterCreated = true
		(*s)[key] = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      name,
			},
			label.Name,
		)
	}
	return newCounterCreated
}

// a previously registered descriptor with the same fully-qualified name as Desc{fqName: "ea_ronny", help: "", constLabels: {}, variableLabels: [market seller]} has different label names or a different help string
// crash wenn change tag count
// to do catch error

func (s *CounterStore) inc(ns string, name string, label ConstLabel, step float64) {
	key := buildKey(ns, name, label)
	if _, ok := (*s)[key]; ok {
		(*s)[key].WithLabelValues(label.Value...).Add(step)
	}
}

func buildKey(ns string, name string, label ConstLabel) string {
	return ns + "_" + name + "__" + strings.Join(label.Name, "_")
}
