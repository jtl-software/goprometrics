package main

import (
	"encoding/json"
	"github.com/prometheus/client_golang/prometheus"
	"io/ioutil"
	"net/http"
)

type CounterStore map[string]prometheus.Counter

var Store CounterStore

func createNewCounterOpts(name string, request *http.Request) (counter prometheus.CounterOpts, success bool) {

	b, err := ioutil.ReadAll(request.Body)
	if err != nil || len(b) == 0 {
		nc := createSimpleCounter(name)
		return nc, true
	}

	nc, ok := createCounterOptsFromRequest(b)
	if !ok {
		return nc, ok
	}
	nc.Name = name
	return nc, true
}

func createCounterOptsFromRequest(b []byte) (counter prometheus.CounterOpts, success bool) {
	var req struct {
		Namespace string            `json:"ns"`
		Help      string            `json:"help"`
		Labels    map[string]string `json:"labels"`
	}

	err := json.Unmarshal(b, &req)
	if err != nil {
		return prometheus.CounterOpts{}, false
	}

	opts := prometheus.CounterOpts{
		Namespace:   req.Namespace,
		Help:        req.Help,
		ConstLabels: req.Labels,
	}
	return opts, true
}

func createSimpleCounter(name string) prometheus.CounterOpts {
	return prometheus.CounterOpts{
		Name: name,
	}
}
