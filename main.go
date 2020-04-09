package main

import (
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
	"strconv"
)

func main() {
	Store = CounterStore{}

	go func() {
		m := mux.NewRouter()
		log.Infoln("metrics are getting exposed on :9112")
		m.Path("/").Handler(promhttp.Handler())
		_ = http.ListenAndServe(":9112", m)
	}()

	r := mux.NewRouter()
	log.Infoln("Start Server on :9111")
	r.HandleFunc("/counter/inc/{ns}/{name}/{labels}", incLabeledCounter).Methods("PUT")

	_ = http.ListenAndServe(":9111", r)
}

func incLabeledCounter(writer http.ResponseWriter, request *http.Request) {
	v := mux.Vars(request)
	name := v["name"]
	ns := v["ns"]
	label := createLabels(v["labels"])
	stepWidth := parseStepWidth(request)

	status := http.StatusOK
	if _, ok := Store[name]; !ok {
		log.Infof("New counter %s_%s registered", ns, name)

		status = http.StatusCreated
		Store[name] = promauto.NewCounterVec(
			prometheus.CounterOpts{
				Namespace: ns,
				Name:      name,
			},
			label.Name,
		)
	}

	Store[name].WithLabelValues(label.Value...).Add(stepWidth)
	writer.WriteHeader(status)
}

func parseStepWidth(request *http.Request) float64 {
	inc, _ := strconv.ParseFloat(request.URL.Query().Get("add"), 64)
	if inc <= 0 {
		inc = 1
	}
	return inc
}

