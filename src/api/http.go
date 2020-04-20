package api

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/common/log"
	"goprometrics/src/store"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

var (
	IncCounter = promauto.NewCounter(prometheus.CounterOpts{
		Namespace: "goprometrics",
		Name:      "metric_incremented",
		Help:      "Count when a new metric is incremented or observed",
	})

	DurationMetricReqHandling = promauto.NewHistogram(prometheus.HistogramOpts{
		Namespace: "goprometrics",
		Name:      "metric_enlarge_duration_seconds",
		Help:      "Duration of metric handling during http request",
		Buckets: []float64{
			0.000001, // 1 Microsecond
			0.0001,   // 0.1 Millisecond
			0.001,    // 1 Millisecond
			0.01,
			0.1,
			1.0, // 1 Second
		},
	})
)

type adapter struct {
	r          *mux.Router
	config     HostConfig
	httpServer *http.Server
}

func NewAdapter(config HostConfig) *adapter {
	return &adapter{r: mux.NewRouter(), config: config}
}

func (a adapter) CounterHandleFunc(h func(writer http.ResponseWriter, request *http.Request)) {
	a.r.HandleFunc("/count/{ns}/{name}", h).Methods("PUT")
}

func (a adapter) SummaryHandleFunc(h func(writer http.ResponseWriter, request *http.Request)) {
	a.r.HandleFunc("/sum/{ns}/{name}/{observation:[0-9]*\\.?[0-9]+}", h).Methods("PUT")
}

func (a adapter) HistogramHandleFunc(h func(writer http.ResponseWriter, request *http.Request)) {
	a.r.HandleFunc("/observe/{ns}/{name}/{observation:[0-9]*\\.?[0-9]+}", h).Methods("PUT")
}

func (a adapter) Serve() {
	log.Infof("Start Server on %s:%s", a.config.host, a.config.port)
	a.listenAndServe()
}

func (a adapter) ServeMetrics() {
	a.r.Path("/metrics").Handler(promhttp.Handler()).Methods("GET")

	log.Infof("Start Metrics Server on %s:%s", a.config.host, a.config.port)
	a.listenAndServe()
}

func (a adapter) listenAndServe() {
	a.httpServer = &http.Server{
		Addr:    a.config.host + ":" + a.config.port,
		Handler: a.r,
	}
	err := a.httpServer.ListenAndServe()
	if err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}

func (a adapter) RequestHandler(s store.Store) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()

		v := mux.Vars(r)
		opts, value, err := createPrometheusMetricOpts(r, v)
		if err != nil {
			handleBadRequestError(err, w)
			return
		}

		created, err := store.Append(s, opts)
		if err != nil {
			handleBadRequestError(err, w)
			return
		}
		s.Inc(opts, value)
		IncCounter.Inc()

		handleResponse(created, w)
		DurationMetricReqHandling.Observe(time.Since(start).Seconds())
	}
}

func createPrometheusMetricOpts(r *http.Request, v map[string]string) (opts store.MetricOpts, value float64, err error) {
	opts.Ns = v["ns"]
	opts.Name = v["name"]

	_ = r.ParseForm()
	opts.Label = createLabels(r.FormValue("labels"))
	opts.SummaryObjectives = parseObjectives(r.FormValue("objectives"))
	opts.HistogramBuckets = parseBuckets(r.FormValue("buckets"))
	opts.Help = r.FormValue("help")

	if v, ok := v["observation"]; ok {
		formPath, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return opts, value, err
		}
		value = formPath
	} else {
		value = parseStepWidth(r)
	}

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

func createLabels(fromRequest string) store.ConstLabel {
	var l store.ConstLabel

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
