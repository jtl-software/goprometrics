package main

import (
	"flag"
)

type ConstLabel struct {
	Name  []string
	Value []string
}

type Counter interface {
	inc(ns string, name string, label ConstLabel, step float64, help string) (bool, error)
}

type PrometheusMetrics struct {
	counterStore CounterStore
}

func main() {
	var port = flag.String("port", "9111", "Port for incoming API requests")
	var portMetrics = flag.String("portm", "9112", "Port for prometheus scraping")
	flag.Parse()

	counter := PrometheusMetrics{counterStore: CounterStore{}}

	go func() {
		NewAdapter().ServeMetrics(*portMetrics)
	}()

	api := NewAdapter()
	handler := api.MakeCounterHandler(counter)
	api.CounterHandleFunc(handler)
	api.Serve(*port)
}

func (pm PrometheusMetrics) inc(ns string, name string, label ConstLabel, step float64, help string) (bool, error) {
	created, err := pm.counterStore.addCounter(ns, name, label, help)
	if err == nil {
		pm.counterStore.inc(ns, name, label, step)
	}
	return created, err
}
