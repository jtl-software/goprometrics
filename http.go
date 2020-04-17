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
	a.r.HandleFunc("/sum/{ns}/{name}/{observation:[0-9]*\\.?[0-9]+}", h).Methods("PUT")
}

func (a Adapter) HistogramHandleFunc(h func(writer http.ResponseWriter, request *http.Request)) {
	a.r.HandleFunc("/observe/{ns}/{name}/{observation:[0-9]*\\.?[0-9]+}", h).Methods("PUT")
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

func (a Adapter) RequestHandler(
	s MetricStore,
	creator func(opts PrometheusMetricOpts, s *MetricStore),
	enlarge func(s *MetricStore, opts PrometheusMetricOpts, value float64),
) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		v := mux.Vars(r)
		opts := createPrometheusMetricOpts(r, v)

		var value float64
		if v, ok := v["observation"]; ok {
			formPath, err := strconv.ParseFloat(v, 64)
			if err != nil {
				handleBadRequestError(err, w)
				return
			}
			value = formPath
		} else {
			value = parseStepWidth(r)
		}

		created, err := s.append(opts, creator)
		if err != nil {
			handleBadRequestError(err, w)
			return
		}

		if s.has(opts) {
			enlarge(&s, opts, value)
		}
		handleResponse(created, w)
	}
}

func createPrometheusMetricOpts(r *http.Request, v map[string]string) (opts PrometheusMetricOpts) {
	opts.ns = v["ns"]
	opts.name = v["name"]

	_ = r.ParseForm()
	opts.label = createLabels(r.FormValue("labels"))
	opts.summaryObjectives = parseObjectives(r.FormValue("objectives"))
	opts.histogramBuckets = parseBuckets(r.FormValue("buckets"))
	opts.help = r.FormValue("help")

	return
}

func parseBuckets(s string) []float64 {
	obj := strings.Split(s, ",")

	buckets := make([]float64, 0)
	for _, value := range obj {
		bucket, err := strconv.ParseFloat(value, 64)
		if err != nil {
			continue
		}
		buckets = append(buckets, bucket)
	}

	sort.Float64s(buckets)
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
