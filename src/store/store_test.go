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
