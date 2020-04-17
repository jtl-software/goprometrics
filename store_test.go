package main

import (
	"errors"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"reflect"
	"testing"
)

func TestPrometheusMetricOpts_BuildStoreKey(t *testing.T) {
	type fields struct {
		ns                string
		name              string
		label             ConstLabel
		help              string
		histogramBuckets  []float64
		summaryObjectives map[float64]float64
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name:   "w/o labels",
			fields: fields{ns: "foo", name: "bar"},
			want:   "foo_bar__",
		},
		{
			name:   "w labels",
			fields: fields{ns: "foo", name: "bar", label: ConstLabel{Name: []string{"a", "b"}}},
			want:   "foo_bar__a_b",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := PrometheusMetricOpts{
				ns:                tt.fields.ns,
				name:              tt.fields.name,
				label:             tt.fields.label,
				help:              tt.fields.help,
				histogramBuckets:  tt.fields.histogramBuckets,
				summaryObjectives: tt.fields.summaryObjectives,
			}
			if got := opts.BuildStoreKey(); got != tt.want {
				t.Errorf("BuildStoreKey() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMetricStore_append(t *testing.T) {
	type fields struct {
		CounterStore   map[string]*prometheus.CounterVec
		SummaryStore   map[string]*prometheus.SummaryVec
		HistogramStore map[string]*prometheus.HistogramVec
	}
	type args struct {
		opts    PrometheusMetricOpts
		creator func(opts PrometheusMetricOpts, s *MetricStore)
	}
	tests := []struct {
		name           string
		fields         fields
		args           args
		wantNewCreated bool
		wantErr        bool
	}{
		{
			name: "create-new-counter",
			fields: fields{
				CounterStore:   map[string]*prometheus.CounterVec{},
				SummaryStore:   map[string]*prometheus.SummaryVec{},
				HistogramStore: map[string]*prometheus.HistogramVec{},
			},
			args: args{
				opts:    PrometheusMetricOpts{name: "foo"},
				creator: CreateCounterMetricHandler(),
			},
			wantNewCreated: true,
			wantErr:        false,
		},
		{
			name: "already-exists-no-new-created",
			fields: fields{
				CounterStore: map[string]*prometheus.CounterVec{
					"_foo__": prometheus.NewCounterVec(prometheus.CounterOpts{Name: "egal"}, []string{}),
				},
				SummaryStore:   map[string]*prometheus.SummaryVec{},
				HistogramStore: map[string]*prometheus.HistogramVec{},
			},
			args: args{
				opts:    PrometheusMetricOpts{name: "foo"},
				creator: CreateSummaryMetricHandler(),
			},
			wantNewCreated: false,
			wantErr:        false,
		},
		{
			name: "can-return-error",
			fields: fields{
				CounterStore:   map[string]*prometheus.CounterVec{},
				SummaryStore:   map[string]*prometheus.SummaryVec{},
				HistogramStore: map[string]*prometheus.HistogramVec{},
			},
			args: args{
				opts: PrometheusMetricOpts{name: "foo"},
				creator: func(opts PrometheusMetricOpts, s *MetricStore) {
					err := errors.New("We expect to have an error")
					panic(err)
				},
			},
			wantNewCreated: false,
			wantErr:        true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MetricStore{
				CounterStore:   tt.fields.CounterStore,
				SummaryStore:   tt.fields.SummaryStore,
				HistogramStore: tt.fields.HistogramStore,
			}
			gotNewCreated, err := s.append(tt.args.opts, tt.args.creator)
			if (err != nil) != tt.wantErr {
				t.Errorf("append() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if gotNewCreated != tt.wantNewCreated {
				t.Errorf("append() gotNewCreated = %v, want %v", gotNewCreated, tt.wantNewCreated)
			}
		})
	}
}

func TestMetricStore_has(t *testing.T) {
	type fields struct {
		CounterStore   map[string]*prometheus.CounterVec
		SummaryStore   map[string]*prometheus.SummaryVec
		HistogramStore map[string]*prometheus.HistogramVec
	}
	type args struct {
		opts PrometheusMetricOpts
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   bool
	}{
		{
			name: "no-match",
			fields: fields{
				CounterStore:   map[string]*prometheus.CounterVec{},
				SummaryStore:   map[string]*prometheus.SummaryVec{},
				HistogramStore: map[string]*prometheus.HistogramVec{},
			},
			args: args{PrometheusMetricOpts{
				ns:   "foo",
				name: "bar",
			}},
			want: false,
		},
		{
			name: "with-match",
			fields: fields{
				CounterStore: map[string]*prometheus.CounterVec{
					"foo_bar__": prometheus.NewCounterVec(prometheus.CounterOpts{Name: "egal"}, []string{}),
				},
				SummaryStore:   map[string]*prometheus.SummaryVec{},
				HistogramStore: map[string]*prometheus.HistogramVec{},
			},
			args: args{PrometheusMetricOpts{
				ns:   "foo",
				name: "bar",
			}},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &MetricStore{
				CounterStore:   tt.fields.CounterStore,
				SummaryStore:   tt.fields.SummaryStore,
				HistogramStore: tt.fields.HistogramStore,
			}
			if got := s.has(tt.args.opts); got != tt.want {
				t.Errorf("has() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewMetricStore(t *testing.T) {
	tests := []struct {
		name string
		want MetricStore
	}{
		{
			name: "Expect empty stores",
			want: MetricStore{
				CounterStore:   map[string]*prometheus.CounterVec{},
				SummaryStore:   map[string]*prometheus.SummaryVec{},
				HistogramStore: map[string]*prometheus.HistogramVec{},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMetricStore(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMetricStore() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestObserveHistogramHandler(t *testing.T) {
	store := NewMetricStore()
	opts := PrometheusMetricOpts{name: "TestObserveHistogramHandler"}
	_, _ = store.append(opts, CreateHistogramMetricHandler())

	ObserveHistogramHandler()(&store, opts, 2.5)
	ObserveHistogramHandler()(&store, opts, 2.5)
	ObserveHistogramHandler()(&store, opts, 2.5)

	if false {
		t.Error("I have no idea how to check results of a Histogram metric - no panic? - fin")
	}
}

func TestObserveSummaryHandler(t *testing.T) {
	store := NewMetricStore()
	opts := PrometheusMetricOpts{name: "TestObserveSummaryHandler"}
	_, _ = store.append(opts, CreateSummaryMetricHandler())

	ObserveSummaryHandler()(&store, opts, 2.5)
	ObserveSummaryHandler()(&store, opts, 2.5)
	ObserveSummaryHandler()(&store, opts, 2.5)

	if false {
		t.Error("I have no idea how to check results of a Summary metric - no panic? - fin")
	}
}
func TestIncCounterHandler(t *testing.T) {
	store := NewMetricStore()
	opts := PrometheusMetricOpts{name: "TestIncCounterHandler"}
	_, _ = store.append(opts, CreateCounterMetricHandler())

	IncCounterHandler()(&store, opts, 1)
	IncCounterHandler()(&store, opts, 2.5)

	result := testutil.ToFloat64(store.CounterStore[opts.BuildStoreKey()])

	if result != 3.5 {
		t.Errorf("Expect to have a counter of 3.5 - actual %f", result)
	}
}

func TestCreateCounterMetricHandler(t *testing.T) {
	store := NewMetricStore()

	creator := CreateCounterMetricHandler()
	creator(PrometheusMetricOpts{name: "TestCreateCounterMetricHandler"}, &store)

	if len(store.CounterStore) != 1 {
		t.Errorf("Expect to have a new counter in store  - actual %d", len(store.CounterStore))
	}
}

func TestCreateHistogramMetricHandler(t *testing.T) {
	store := NewMetricStore()

	creator := CreateHistogramMetricHandler()
	creator(PrometheusMetricOpts{name: "TestCreateHistogramMetricHandler"}, &store)

	if len(store.HistogramStore) != 1 {
		t.Errorf("Expect to have a new histogram in store  - actual %d", len(store.CounterStore))
	}
}

func TestCreateSummaryMetricHandler(t *testing.T) {
	store := NewMetricStore()

	creator := CreateSummaryMetricHandler()
	creator(PrometheusMetricOpts{name: "TestCreateSummaryMetricHandler"}, &store)

	if len(store.SummaryStore) != 1 {
		t.Errorf("Expect to have a new summary in store  - actual %d", len(store.CounterStore))
	}
}

func TestCanCreateNewMetricStore(t *testing.T) {
	store := NewMetricStore()

	if store.CounterStore == nil {
		t.Errorf("Expect to have a store.CounterStore - nil given")
	}

	if store.SummaryStore == nil {
		t.Errorf("Expect to have a store.SummaryStore - nil given")
	}

	if store.HistogramStore == nil {
		t.Errorf("Expect to have a store.SummaryStore - nil given")
	}
}
