package main

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
)

type Adapter struct {
	r      *mux.Router
	config HostConfig
}

func NewAdapter(config HostConfig) Adapter {
	return Adapter{r: mux.NewRouter(), config: config}
}

func (a Adapter) CounterHandleFunc(h func(writer http.ResponseWriter, request *http.Request)) {
	a.r.HandleFunc("/count/{ns}/{name}", h).Methods("PUT")
}

func (a Adapter) SummaryHandleFunc(h func(writer http.ResponseWriter, request *http.Request)) {
	a.r.HandleFunc("/sum/{ns}/{name}/{observation}", h).Methods("PUT")
}

func (a Adapter) HistogramHandleFunc(h func(writer http.ResponseWriter, request *http.Request)) {
	a.r.HandleFunc("/observe/{ns}/{name}/{observation}", h).Methods("PUT")
}

func (a Adapter) Serve() {
	log.Infof("Start Server on %s:%s", a.config.host, a.config.port)
	a.listenAndServe()
}

func (a Adapter) ServeMetrics() {
	a.r.Path("/metrics").Handler(promhttp.Handler())

	log.Infof("Start Metrics Server on %s:%s", a.config.host, a.config.port)
	a.listenAndServe()
}

func (a Adapter) listenAndServe() {
	err := http.ListenAndServe(a.config.host+":"+a.config.port, a.r)
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func (a Adapter) MakeCounterHandler(counter PrometheusCounter) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		v := mux.Vars(request)
		name := v["name"]
		ns := v["ns"]
		step := parseStepWidth(request)

		_ = request.ParseForm()
		label := createLabels(request.FormValue("labels"))
		help := request.FormValue("help")

		created, err := counter.incCounter(ns, name, label, step, help)
		if err == nil {
			handleResponse(created, writer)
		} else {
			handleBadRequestError(err, writer)
		}
	}
}

func (a Adapter) MakeSummaryHandler(summary PrometheusSummary) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		v := mux.Vars(request)
		name := v["name"]
		ns := v["ns"]
		observation, err := strconv.ParseFloat(v["observation"], 64)
		if err != nil {
			handleBadRequestError(err, writer)
			return
		}

		_ = request.ParseForm()
		label := createLabels(request.FormValue("labels"))
		objectives := parseObjectives(request.FormValue("objectives"))
		help := request.FormValue("help")

		created, err := summary.sum(ns, name, label, observation, objectives, help)
		if err == nil {
			handleResponse(created, writer)
		} else {
			handleBadRequestError(err, writer)
		}
	}
}

func (a Adapter) MakeHistogramHandler(histogram PrometheusHistogram) func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {
		v := mux.Vars(request)
		name := v["name"]
		ns := v["ns"]
		observation, err := strconv.ParseFloat(v["observation"], 64)
		if err != nil {
			handleBadRequestError(err, writer)
			return
		}

		_ = request.ParseForm()
		label := createLabels(request.FormValue("labels"))
		buckets := parseBuckets(request.FormValue("buckets"))
		help := request.FormValue("help")

		created, err := histogram.observe(ns, name, label, observation, buckets, help)
		if err == nil {
			handleResponse(created, writer)
		} else {
			handleBadRequestError(err, writer)
		}
	}
}

func parseBuckets(s string) (buckets []float64) {
	obj := strings.Split(s, ",")
	sort.Strings(obj)

	for _, value := range obj {
		bucket, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}
		buckets = append(buckets, bucket)
	}

	return buckets
}

func parseObjectives(s string) map[float64]float64 {

	objectives := make(map[float64]float64, 0)
	if s == "" {
		return objectives
	}

	obj := strings.Split(s, ",")
	sort.Strings(obj)

	for _, value := range obj {
		parts := strings.Split(value, ":")

		key, err := strconv.ParseFloat(parts[0], 64)
		if err != nil {
			continue
		}

		value, err := strconv.ParseFloat(parts[1], 64)
		if err != nil {
			continue
		}

		objectives[key] = value
	}

	return objectives
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
