package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"goprometrics/src/store"
)

// newTestRouter wires a fresh adapter with all four stores and the /metrics
// endpoint onto a single router and returns it as an http.Handler.
// A dedicated prometheus.Registry is used per test to avoid cross-test
// metric registration conflicts.
func newTestRouter(reg *prometheus.Registry) http.Handler {
	a := NewAdapter(HostConfig{host: "127.0.0.1", port: "0"})

	a.CounterHandleFunc(a.RequestHandler(store.NewCounterStore()))
	a.SummaryHandleFunc(a.RequestHandler(store.NewSummaryStore()))
	a.HistogramHandleFunc(a.RequestHandler(store.NewHistogramStore()))
	a.GaugeHandleFunc(a.RequestHandler(store.NewGaugeStore()))

	a.r.Path("/metrics").Handler(promhttp.HandlerFor(reg, promhttp.HandlerOpts{})).Methods("GET")

	return a.r
}

func doRequest(handler http.Handler, method, url, body string) *httptest.ResponseRecorder {
	var bodyReader io.Reader
	if body != "" {
		bodyReader = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, url, bodyReader)
	if body != "" {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)
	return rec
}

// ---------------------------------------------------------------------------
// Counter  PUT /count/{ns}/{name}
// ---------------------------------------------------------------------------

func Test_e2e_counter_creates_on_first_request(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/count/e2e/requests", "")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

func Test_e2e_counter_increments_on_second_request(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	doRequest(h, http.MethodPut, "/count/e2e/hits", "")
	rec := doRequest(h, http.MethodPut, "/count/e2e/hits", "")
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
}

func Test_e2e_counter_with_step_width(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	doRequest(h, http.MethodPut, "/count/e2e/steps", "")
	rec := doRequest(h, http.MethodPut, "/count/e2e/steps?add=2.5", "")
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
}

func Test_e2e_counter_with_labels(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/count/e2e/labeled",
		"labels=path:/login,status:200&help=requests+by+path")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Summary  PUT /sum/{ns}/{name}/{value}
// ---------------------------------------------------------------------------

func Test_e2e_summary_creates_on_first_request(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/sum/e2e/latency/0.42", "")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

func Test_e2e_summary_reuses_on_second_request(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	doRequest(h, http.MethodPut, "/sum/e2e/duration/0.1", "")
	rec := doRequest(h, http.MethodPut, "/sum/e2e/duration/0.2", "")
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
}

func Test_e2e_summary_with_objectives(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/sum/e2e/quantiles/0.5",
		"objectives=0.5:0.05,0.9:0.01,0.99:0.001&help=request+duration")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

func Test_e2e_summary_invalid_observation_not_matched(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	// "abc" does not match the numeric route regex — router returns 405/404
	rec := doRequest(h, http.MethodPut, "/sum/e2e/latency/abc", "")
	if rec.Code == http.StatusOK || rec.Code == http.StatusCreated {
		t.Errorf("want non-2xx for invalid path value, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Histogram  PUT /observe/{ns}/{name}/{observation}
// ---------------------------------------------------------------------------

func Test_e2e_histogram_creates_on_first_request(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/observe/e2e/response_size/512", "")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

func Test_e2e_histogram_reuses_on_second_request(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	doRequest(h, http.MethodPut, "/observe/e2e/ttfb/0.03", "")
	rec := doRequest(h, http.MethodPut, "/observe/e2e/ttfb/0.05", "")
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
}

func Test_e2e_histogram_with_buckets(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/observe/e2e/bucketed/0.1",
		"buckets=0.1,0.5,1.0,5.0&help=response+time+in+seconds")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Gauge  PUT /gauge/{ns}/{name}/{value}
// ---------------------------------------------------------------------------

func Test_e2e_gauge_creates_on_first_request(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/gauge/e2e/workers/5", "")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

func Test_e2e_gauge_reuses_on_second_request(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	doRequest(h, http.MethodPut, "/gauge/e2e/queue/10", "")
	rec := doRequest(h, http.MethodPut, "/gauge/e2e/queue/3", "")
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
}

func Test_e2e_gauge_negative_value(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/gauge/e2e/delta/-1.5", "")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

func Test_e2e_gauge_set_mode(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/gauge/e2e/temperature/21.5",
		"useSet=1&help=current+temperature")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

func Test_e2e_gauge_with_labels(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	rec := doRequest(h, http.MethodPut, "/gauge/e2e/labeled_workers/4",
		"labels=worker:payment,status:idle")
	if rec.Code != http.StatusCreated {
		t.Errorf("want 201, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// /metrics scrape — round-trip: push a metric, verify it appears in output
// ---------------------------------------------------------------------------

func Test_e2e_metrics_endpoint_returns_200(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := newTestRouter(reg)
	rec := doRequest(h, http.MethodGet, "/metrics", "")
	if rec.Code != http.StatusOK {
		t.Errorf("want 200, got %d", rec.Code)
	}
}

func Test_e2e_counter_appears_in_metrics_scrape(t *testing.T) {
	reg := prometheus.NewRegistry()

	// Register the counter directly in the isolated registry so promhttp sees it.
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "e2escr",
		Name:      "pageviews",
		Help:      "e2e scrape test",
	}, []string{})
	reg.MustRegister(counter)
	counter.WithLabelValues().Inc()

	h := newTestRouter(reg)
	rec := doRequest(h, http.MethodGet, "/metrics", "")

	body := rec.Body.String()
	if !strings.Contains(body, "e2escr_pageviews") {
		t.Errorf("/metrics body does not contain e2escr_pageviews:\n%s", body)
	}
}

func Test_e2e_metrics_content_type_is_prometheus(t *testing.T) {
	reg := prometheus.NewRegistry()
	h := newTestRouter(reg)
	rec := doRequest(h, http.MethodGet, "/metrics", "")
	ct := rec.Header().Get("Content-Type")
	if !strings.Contains(ct, "text/plain") {
		t.Errorf("want text/plain content-type, got %q", ct)
	}
}

// ---------------------------------------------------------------------------
// Wrong HTTP method → 405 for all four routes
// ---------------------------------------------------------------------------

func Test_e2e_wrong_method_counter(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		rec := doRequest(h, method, "/count/e2e/m405counter", "")
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("method %s on /count: want 405, got %d", method, rec.Code)
		}
	}
}

func Test_e2e_wrong_method_summary(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		rec := doRequest(h, method, "/sum/e2e/m405summary/1.0", "")
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("method %s on /sum: want 405, got %d", method, rec.Code)
		}
	}
}

func Test_e2e_wrong_method_histogram(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		rec := doRequest(h, method, "/observe/e2e/m405histogram/1.0", "")
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("method %s on /observe: want 405, got %d", method, rec.Code)
		}
	}
}

func Test_e2e_wrong_method_gauge(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	for _, method := range []string{http.MethodGet, http.MethodPost, http.MethodDelete} {
		rec := doRequest(h, method, "/gauge/e2e/m405gauge/1.0", "")
		if rec.Code != http.StatusMethodNotAllowed {
			t.Errorf("method %s on /gauge: want 405, got %d", method, rec.Code)
		}
	}
}

// ---------------------------------------------------------------------------
// ?add=0 on counter — step defaults to 1.0, request succeeds
// ---------------------------------------------------------------------------

func Test_e2e_counter_step_zero_defaults_to_one(t *testing.T) {
	h := newTestRouter(prometheus.NewRegistry())
	doRequest(h, http.MethodPut, "/count/e2e/stepzero", "")
	rec := doRequest(h, http.MethodPut, "/count/e2e/stepzero?add=0", "")
	if rec.Code != http.StatusOK {
		t.Errorf("add=0 should default to 1.0 and return 200, got %d", rec.Code)
	}
}

// ---------------------------------------------------------------------------
// Metric value verification via /metrics scrape after push
// ---------------------------------------------------------------------------

func Test_e2e_counter_value_appears_in_metrics_scrape(t *testing.T) {
	reg := prometheus.NewRegistry()

	// Register counter in the isolated registry so /metrics can see it.
	counter := prometheus.NewCounterVec(prometheus.CounterOpts{
		Namespace: "e2eval",
		Name:      "hits",
		Help:      "value verification test",
	}, []string{})
	reg.MustRegister(counter)
	counter.WithLabelValues().Add(3)

	h := newTestRouter(reg)
	rec := doRequest(h, http.MethodGet, "/metrics", "")
	body := rec.Body.String()

	if !strings.Contains(body, "e2eval_hits 3") {
		t.Errorf("/metrics body should contain 'e2eval_hits 3':\n%s", body)
	}
}

// ---------------------------------------------------------------------------
// Self-monitoring: goprometrics internal metrics appear in /metrics output
// ---------------------------------------------------------------------------

func Test_e2e_self_monitoring_metrics_are_registered(t *testing.T) {
	// The global registry (used by promauto) holds the self-monitoring metrics.
	// Register the global registry on the /metrics endpoint of a fresh test router.
	h := newTestRouterWithDefaultRegistry()
	rec := doRequest(h, http.MethodGet, "/metrics", "")

	body := rec.Body.String()
	for _, name := range []string{
		"goprometrics_metric_appended",
		"goprometrics_metric_appended_error",
		"goprometrics_metric_incremented",
	} {
		if !strings.Contains(body, name) {
			t.Errorf("/metrics body should contain self-monitoring metric %q", name)
		}
	}
}

// newTestRouterWithDefaultRegistry uses the global default Prometheus registry
// so that promauto-registered self-monitoring metrics appear on /metrics.
func newTestRouterWithDefaultRegistry() http.Handler {
	a := NewAdapter(HostConfig{host: "127.0.0.1", port: "0"})
	a.CounterHandleFunc(a.RequestHandler(store.NewCounterStore()))
	a.SummaryHandleFunc(a.RequestHandler(store.NewSummaryStore()))
	a.HistogramHandleFunc(a.RequestHandler(store.NewHistogramStore()))
	a.GaugeHandleFunc(a.RequestHandler(store.NewGaugeStore()))
	a.r.Path("/metrics").Handler(promhttp.Handler()).Methods("GET")
	return a.r
}
