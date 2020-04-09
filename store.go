package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"sort"
	"strings"
)

type CounterStore map[string]*prometheus.CounterVec

var Store CounterStore

type ConstLabel struct {
	Name []string
	Value []string
}

func createLabels(fromRequest string) ConstLabel {
	var l ConstLabel

	labels := strings.Split(fromRequest, ",")
	sort.Strings(labels)

	for _, value := range labels {
		parts := strings.Split(value, ":")
		if len(parts) == 2 {
			l.Name = append(l.Name, parts[0])
			l.Value = append(l.Value, parts[1])
		}
	}
	return l
}
