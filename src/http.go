package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
	"sort"
	"strconv"
	"strings"
)

type Adapter struct {
	r *mux.Router
}

func NewAdapter() Adapter {
	return Adapter{r: mux.NewRouter()}
}

func (a Adapter) CounterHandleFunc(h func(writer http.ResponseWriter, request *http.Request)) {
	a.r.HandleFunc("/count/{ns}/{name}/{labels}", h).Methods("PUT")
}

func (a Adapter) Serve() {
	log.Infof("Start Server on :9111")
	_ = http.ListenAndServe(":9111", a.r)
}

func (a Adapter) ServeMetrics() {
	a.r.Path("/metrics").Handler(promhttp.Handler())

	log.Infoln("metrics are getting exposed on :9112")
	_ = http.ListenAndServe(":9112", a.r)
}

func (a Adapter) MakeCounterHandler(counter Counter) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		v := mux.Vars(request)
		name := v["name"]
		ns := v["ns"]
		label := createLabels(v["labels"])
		step := parseStepWidth(request)

		created, err := counter.inc(ns, name, label, step)
		if err == nil {
			handleResponse(created, writer)
		} else {
			handleBadRequestError(err, writer)
		}
	}
}

func handleResponse(created bool, writer http.ResponseWriter) {
	if created == true {
		writer.WriteHeader(http.StatusCreated)
	} else {
		writer.WriteHeader(http.StatusOK)
	}
}

func handleBadRequestError(err error, writer http.ResponseWriter) {
	b, _ := json.Marshal(struct {
		Message string `json:"message"`
	}{
		Message: err.Error(),
	})

	writer.WriteHeader(http.StatusBadRequest)
	_, _ = writer.Write(b)
}

func parseStepWidth(request *http.Request) float64 {
	inc, _ := strconv.ParseFloat(request.URL.Query().Get("add"), 64)
	if inc <= 0 {
		inc = 1
	}
	return inc
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