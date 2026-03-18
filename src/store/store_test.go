package store

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"reflect"
	"testing"
)

func TestNewCounterStore(t *testing.T) {
	tests := []struct {
		name string
		want Store
	}{
		{
			name: "can create a new empty Counter Store",
			want: counterStore{map[string]*prometheus.CounterVec{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewCounterStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewCounterStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewHistogramStore(t *testing.T) {
	tests := []struct {
		name string
		want Store
	}{
		{
			name: "can create a new empty Histogram Store",
			want: histogramStore{map[string]*prometheus.HistogramVec{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewHistogramStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewHistogramStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewSummaryStore(t *testing.T) {
	tests := []struct {
		name string
		want Store
	}{
		{
			name: "can create a new empty Summary Store",
			want: summaryStore{map[string]*prometheus.SummaryVec{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewSummaryStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewSummaryStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewGaugeStore(t *testing.T) {
	tests := []struct {
		name string
		want Store
	}{
		{
			name: "can create a new empty Gauge Store",
			want: gaugeStore{map[string]*prometheus.GaugeVec{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewGaugeStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewGaugeStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_counterStore_Append(t *testing.T) {
	store := counterStore{map[string]*prometheus.CounterVec{}}
	store.Append(MetricOpts{Name: "Test_counterStore_Append"})
	if len(store.store) != 1 {
		t.Errorf("Expect to have exactly 1 counter in store - %d given", len(store.store))
	}

}

func Test_counterStore_Has(t *testing.T) {
	store := counterStore{map[string]*prometheus.CounterVec{}}
	opts := MetricOpts{Name: "Test_counterStore_Has"}
	store.Append(opts)
	if store.Has(opts) != true {
		t.Errorf("Expect to have a Item for Key %s in Counter Store", opts.Key())
	}
}

func Test_counterStore_Inc(t *testing.T) {
	store := counterStore{map[string]*prometheus.CounterVec{}}
	opts := MetricOpts{Name: "Test_counterStore_Inc"}
	store.Append(opts)
	store.Inc(opts, 2.0)

	result := testutil.ToFloat64(store.store[opts.Key()])
	if result != 2.0 {
		t.Errorf("Expect to have a Counter in state 2.0 - result %f", result)
	}
}

func Test_histogramStore_Append(t *testing.T) {
	store := histogramStore{map[string]*prometheus.HistogramVec{}}
	store.Append(MetricOpts{Name: "Test"})
	if len(store.store) != 1 {
		t.Errorf("Expect to have exactly 1 histogram in store - %d given", len(store.store))
	}
}

func Test_histogramStore_Has(t *testing.T) {
	store := histogramStore{map[string]*prometheus.HistogramVec{}}
	opts := MetricOpts{Name: "Test_histogramStore_Has"}
	store.Append(opts)
	if store.Has(opts) != true {
		t.Errorf("Expect to have a Item for Key %s in HistogramStore Store", opts.Key())
	}
}

func Test_histogramStore_Inc(t *testing.T) {
	store := histogramStore{map[string]*prometheus.HistogramVec{}}
	opts := MetricOpts{Name: "Test_histogramStore_Inc"}
	store.Append(opts)
	store.Inc(opts, 3.0)

	// it is a bit overcomplicated to check the current results for our metric
	// so we do it shot hand and just check if the store has anything which
	// can by collected
	result := testutil.CollectAndCount(store.store[opts.Key()])
	if result != 1 {
		t.Errorf("Expect to have a Summary which can be colleted - result %d", result)
	}
}

func Test_summaryStore_Append(t *testing.T) {
	store := summaryStore{map[string]*prometheus.SummaryVec{}}
	store.Append(MetricOpts{Name: "Test_summaryStore_Append"})
	if len(store.store) != 1 {
		t.Errorf("Expect to have exactly 1 summary in store - %d given", len(store.store))
	}
}

func Test_summaryStore_Has(t *testing.T) {
	store := summaryStore{map[string]*prometheus.SummaryVec{}}
	opts := MetricOpts{Name: "Test_summaryStore_Has"}
	store.Append(opts)
	if store.Has(opts) != true {
		t.Errorf("Expect to have a Item for Key %s in Summary Store", opts.Key())
	}
}

func Test_summaryStore_Inc(t *testing.T) {
	store := summaryStore{map[string]*prometheus.SummaryVec{}}
	opts := MetricOpts{Name: "Test_summaryStore_Inc"}
	store.Append(opts)
	store.Inc(opts, 3.0)

	// it is a bit overcomplicated to check the current results for our metric
	// so we do it shot hand and just check if the store has anything which
	// can by collected
	result := testutil.CollectAndCount(store.store[opts.Key()])
	if result != 1 {
		t.Errorf("Expect to have a Summary which can be colleted - result %d", result)
	}
}

func Test_GaugeStore_Append(t *testing.T) {
	store := gaugeStore{map[string]*prometheus.GaugeVec{}}
	store.Append(MetricOpts{Name: "Test_GaugeStore_Append"})
	if len(store.store) != 1 {
		t.Errorf("Expect to have exactly 1 gauge in store - %d given", len(store.store))
	}
}

func Test_GaugeStore_Inc(t *testing.T) {
	store := gaugeStore{map[string]*prometheus.GaugeVec{}}
	opts := MetricOpts{Name: "Test_GaugeStore_Inc"}
	store.Append(opts)
	store.Inc(opts, -1.01)
	store.Inc(opts, 2.02)

	result := testutil.ToFloat64(store.store[opts.Key()])
	if result != 1.01 {
		t.Errorf("Expect to have a Counter in state 1.01 - result %f", result)
	}
}

func Test_GaugeStore_Has(t *testing.T) {
	store := gaugeStore{map[string]*prometheus.GaugeVec{}}
	opts := MetricOpts{Name: "Test_GaugeStore_Has"}
	store.Append(opts)
	if store.Has(opts) != true {
		t.Errorf("Expect to have a Item for Key %s in Gauge Store", opts.Key())
	}
}

// ---------------------------------------------------------------------------
// Has() returns false when metric has not been appended
// ---------------------------------------------------------------------------

func Test_counterStore_Has_returns_false_when_not_present(t *testing.T) {
	s := counterStore{map[string]*prometheus.CounterVec{}}
	opts := MetricOpts{Name: "Test_counterStore_Has_false"}
	if s.Has(opts) != false {
		t.Errorf("Has() should return false for a metric that was never appended")
	}
}

func Test_histogramStore_Has_returns_false_when_not_present(t *testing.T) {
	s := histogramStore{map[string]*prometheus.HistogramVec{}}
	opts := MetricOpts{Name: "Test_histogramStore_Has_false"}
	if s.Has(opts) != false {
		t.Errorf("Has() should return false for a metric that was never appended")
	}
}

func Test_summaryStore_Has_returns_false_when_not_present(t *testing.T) {
	s := summaryStore{map[string]*prometheus.SummaryVec{}}
	opts := MetricOpts{Name: "Test_summaryStore_Has_false"}
	if s.Has(opts) != false {
		t.Errorf("Has() should return false for a metric that was never appended")
	}
}

func Test_GaugeStore_Has_returns_false_when_not_present(t *testing.T) {
	s := gaugeStore{map[string]*prometheus.GaugeVec{}}
	opts := MetricOpts{Name: "Test_GaugeStore_Has_false"}
	if s.Has(opts) != false {
		t.Errorf("Has() should return false for a metric that was never appended")
	}
}

// ---------------------------------------------------------------------------
// Inc() is a no-op when metric has not been appended
// ---------------------------------------------------------------------------

func Test_counterStore_Inc_noop_when_not_appended(t *testing.T) {
	s := counterStore{map[string]*prometheus.CounterVec{}}
	opts := MetricOpts{Name: "Test_counterStore_Inc_noop"}
	// Must not panic; store map stays empty.
	s.Inc(opts, 1.0)
	if len(s.store) != 0 {
		t.Errorf("Inc() on unappended metric should not mutate the store")
	}
}

func Test_histogramStore_Inc_noop_when_not_appended(t *testing.T) {
	s := histogramStore{map[string]*prometheus.HistogramVec{}}
	opts := MetricOpts{Name: "Test_histogramStore_Inc_noop"}
	s.Inc(opts, 1.0)
	if len(s.store) != 0 {
		t.Errorf("Inc() on unappended metric should not mutate the store")
	}
}

func Test_summaryStore_Inc_noop_when_not_appended(t *testing.T) {
	s := summaryStore{map[string]*prometheus.SummaryVec{}}
	opts := MetricOpts{Name: "Test_summaryStore_Inc_noop"}
	s.Inc(opts, 1.0)
	if len(s.store) != 0 {
		t.Errorf("Inc() on unappended metric should not mutate the store")
	}
}

func Test_GaugeStore_Inc_noop_when_not_appended(t *testing.T) {
	s := gaugeStore{map[string]*prometheus.GaugeVec{}}
	opts := MetricOpts{Name: "Test_GaugeStore_Inc_noop"}
	s.Inc(opts, 1.0)
	if len(s.store) != 0 {
		t.Errorf("Inc() on unapprended metric should not mutate the store")
	}
}

// ---------------------------------------------------------------------------
// gaugeStore.Inc with SetGaugeToValue=true uses Set() instead of Add()
// ---------------------------------------------------------------------------

func Test_GaugeStore_Inc_set_mode(t *testing.T) {
	s := gaugeStore{map[string]*prometheus.GaugeVec{}}
	opts := MetricOpts{Name: "Test_GaugeStore_Inc_set_mode"}
	s.Append(opts)

	// Add 10, then Set to 3 — result must be 3, not 13.
	s.Inc(opts, 10.0)
	opts.SetGaugeToValue = true
	s.Inc(opts, 3.0)

	result := testutil.ToFloat64(s.store[opts.Key()])
	if result != 3.0 {
		t.Errorf("Inc() with SetGaugeToValue=true should Set() the gauge; want 3.0, got %f", result)
	}
}

// ---------------------------------------------------------------------------
// Append with labels
// ---------------------------------------------------------------------------

func Test_counterStore_Append_with_labels(t *testing.T) {
	s := counterStore{map[string]*prometheus.CounterVec{}}
	opts := MetricOpts{
		Name:  "Test_counterStore_Append_labels",
		Label: ConstLabel{Name: []string{"env"}, Value: []string{"prod"}},
	}
	s.Append(opts)
	if len(s.store) != 1 {
		t.Errorf("Append with labels should create exactly 1 entry, got %d", len(s.store))
	}
}

func Test_histogramStore_Append_with_buckets(t *testing.T) {
	s := histogramStore{map[string]*prometheus.HistogramVec{}}
	opts := MetricOpts{
		Name:             "Test_histogramStore_Append_buckets",
		HistogramBuckets: []float64{0.1, 0.5, 1.0, 5.0},
	}
	s.Append(opts)
	if len(s.store) != 1 {
		t.Errorf("Append with custom buckets should create exactly 1 entry, got %d", len(s.store))
	}
}

func Test_summaryStore_Append_with_objectives(t *testing.T) {
	s := summaryStore{map[string]*prometheus.SummaryVec{}}
	opts := MetricOpts{
		Name:              "Test_summaryStore_Append_objectives",
		SummaryObjectives: map[float64]float64{0.5: 0.05, 0.9: 0.01},
	}
	s.Append(opts)
	if len(s.store) != 1 {
		t.Errorf("Append with objectives should create exactly 1 entry, got %d", len(s.store))
	}
}

// ---------------------------------------------------------------------------
// Inc observed value verification for histogram and summary
// ---------------------------------------------------------------------------

func Test_histogramStore_Inc_observed_value(t *testing.T) {
	s := histogramStore{map[string]*prometheus.HistogramVec{}}
	opts := MetricOpts{
		Name:             "Test_histogramStore_Inc_value",
		HistogramBuckets: []float64{1.0, 5.0, 10.0},
	}
	s.Append(opts)
	s.Inc(opts, 3.0)

	// Gather via an isolated registry to check SampleSum equals the observed value.
	reg := prometheus.NewRegistry()
	reg.MustRegister(s.store[opts.Key()])
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	if len(families) != 1 || len(families[0].Metric) != 1 {
		t.Fatalf("expected 1 metric family with 1 series, got %v", families)
	}
	got := families[0].Metric[0].Histogram.GetSampleSum()
	if got != 3.0 {
		t.Errorf("histogram SampleSum: want 3.0, got %f", got)
	}
}

func Test_summaryStore_Inc_observed_value(t *testing.T) {
	s := summaryStore{map[string]*prometheus.SummaryVec{}}
	opts := MetricOpts{
		Name:              "Test_summaryStore_Inc_value",
		SummaryObjectives: map[float64]float64{0.5: 0.05},
	}
	s.Append(opts)
	s.Inc(opts, 7.0)

	// Gather via an isolated registry to check SampleSum equals the observed value.
	reg := prometheus.NewRegistry()
	reg.MustRegister(s.store[opts.Key()])
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("gather failed: %v", err)
	}
	if len(families) != 1 || len(families[0].Metric) != 1 {
		t.Fatalf("expected 1 metric family with 1 series, got %v", families)
	}
	got := families[0].Metric[0].Summary.GetSampleSum()
	if got != 7.0 {
		t.Errorf("summary SampleSum: want 7.0, got %f", got)
	}
}
