package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

// func TestAdapter_CounterHandleFunc(t *testing.T) {
// 	type fields struct {
// 		r      *mux.Router
// 		config HostConfig
// 	}
// 	type args struct {
// 		h func(writer http.ResponseWriter, request *http.Request)
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			a := Adapter{
// 				r:      tt.fields.r,
// 				config: tt.fields.config,
// 			}
// 		})
// 	}
// }
//
// func TestAdapter_HistogramHandleFunc(t *testing.T) {
// 	type fields struct {
// 		r      *mux.Router
// 		config HostConfig
// 	}
// 	type args struct {
// 		h func(writer http.ResponseWriter, request *http.Request)
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			a := Adapter{
// 				r:      tt.fields.r,
// 				config: tt.fields.config,
// 			}
// 		})
// 	}
// }
//
type FakeStore struct {
	store map[string]string
}

func (s *FakeStore) has(opts PrometheusMetricOpts) bool {
	return true
}

func (s *FakeStore) append(opts PrometheusMetricOpts, creator func(opts PrometheusMetricOpts, s *MetricStore)) (newCreated bool, err error) {
	return true, nil
}

func TestAdapter_RequestHandler(t *testing.T) {

	creatorCalled := false
	fakeCreator := func(opts PrometheusMetricOpts, s *MetricStore) {
		creatorCalled = true
		s.CounterStore["___"] = promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "beer",
		}, opts.label.Value)
	}

	enlargerCalled := false
	fakeEnlarge := func(s *MetricStore, opts PrometheusMetricOpts, value float64) {
		enlargerCalled = true
	}

	a := NewAdapter(HostConfig{})
	handler := a.RequestHandler(NewMetricStore(), fakeCreator, fakeEnlarge)

	testResponse := httptest.NewRecorder()
	testRequest := httptest.NewRequest("PUT", "/count/{ns}/{name}", nil)
	handler(testResponse, testRequest)

	if !creatorCalled {
		t.Errorf("Expect to have creator called")
	}

	if !enlargerCalled {
		t.Errorf("Expect to have enlarger called")
	}

}

//
// func TestAdapter_Serve(t *testing.T) {
// 	type fields struct {
// 		r      *mux.Router
// 		config HostConfig
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			a := Adapter{
// 				r:      tt.fields.r,
// 				config: tt.fields.config,
// 			}
// 		})
// 	}
// }
//
// func TestAdapter_ServeMetrics(t *testing.T) {
// 	type fields struct {
// 		r      *mux.Router
// 		config HostConfig
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			a := Adapter{
// 				r:      tt.fields.r,
// 				config: tt.fields.config,
// 			}
// 		})
// 	}
// }
//
// func TestAdapter_SummaryHandleFunc(t *testing.T) {
// 	type fields struct {
// 		r      *mux.Router
// 		config HostConfig
// 	}
// 	type args struct {
// 		h func(writer http.ResponseWriter, request *http.Request)
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 		args   args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			a := Adapter{
// 				r:      tt.fields.r,
// 				config: tt.fields.config,
// 			}
// 		})
// 	}
// }
//
// func TestAdapter_listenAndServe(t *testing.T) {
// 	type fields struct {
// 		r      *mux.Router
// 		config HostConfig
// 	}
// 	tests := []struct {
// 		name   string
// 		fields fields
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			a := Adapter{
// 				r:      tt.fields.r,
// 				config: tt.fields.config,
// 			}
// 		})
// 	}
// }
//
// func TestNewAdapter(t *testing.T) {
// 	type args struct {
// 		config HostConfig
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want Adapter
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := NewAdapter(tt.args.config); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("NewAdapter() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//
// func Test_createLabels(t *testing.T) {
// 	type args struct {
// 		fromRequest string
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 		want ConstLabel
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			if got := createLabels(tt.args.fromRequest); !reflect.DeepEqual(got, tt.want) {
// 				t.Errorf("createLabels() = %v, want %v", got, tt.want)
// 			}
// 		})
// 	}
// }
//

func Test_createPrometheusMetricOpts(t *testing.T) {

	request, _ := http.NewRequest("PUT", "foo.de/count/myns/myname", strings.NewReader(""))

	requestWithLabels, _ := http.NewRequest(
		"PUT",
		"foo.de",
		strings.NewReader("labels=foo:bar,beer:wine&objectives=0.5:0.6,0.99:0.10&buckets=1,2&help=beer"))
	requestWithLabels.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	type args struct {
		r *http.Request
		v map[string]string
	}
	tests := []struct {
		name     string
		args     args
		wantOpts PrometheusMetricOpts
	}{
		{
			name: "can parse namespace and name",
			args: args{request, map[string]string{"ns": "myns", "name": "myname"}},
			wantOpts: PrometheusMetricOpts{
				ns:                "myns",
				name:              "myname",
				label:             ConstLabel{},
				help:              "",
				histogramBuckets:  []float64{},
				summaryObjectives: map[float64]float64{},
			},
		},
		{
			name: "can parse from body",
			args: args{requestWithLabels, map[string]string{}},
			wantOpts: PrometheusMetricOpts{
				ns:   "",
				name: "",
				label: ConstLabel{
					Name:  []string{"beer", "foo"},
					Value: []string{"wine", "bar"},
				},
				help:             "beer",
				histogramBuckets: []float64{1, 2},
				summaryObjectives: map[float64]float64{
					0.5:  0.6,
					0.99: 0.10,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotOpts := createPrometheusMetricOpts(tt.args.r, tt.args.v); !reflect.DeepEqual(gotOpts, tt.wantOpts) {
				t.Errorf("createPrometheusMetricOpts() = %v, want %v", gotOpts, tt.wantOpts)
			}
		})
	}
}

//
// func Test_handleBadRequestError(t *testing.T) {
// 	type args struct {
// 		err    error
// 		writer http.ResponseWriter
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 		})
// 	}
// }
//
// func Test_handleResponse(t *testing.T) {
// 	type args struct {
// 		created bool
// 		writer  http.ResponseWriter
// 	}
// 	tests := []struct {
// 		name string
// 		args args
// 	}{
// 		// TODO: Add test cases.
// 	}
// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 		})
// 	}
// }
//
func Test_parseBuckets(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name        string
		args        args
		wantBuckets []float64
	}{
		{
			name:        "can parse buckets",
			args:        args{"5,6.9,10.1,1.0"},
			wantBuckets: []float64{1.0, 5.0, 6.9, 10.1},
		},
		{
			name:        "invalid bucket values are ignored",
			args:        args{"5,6.9,seemsInvalid,bad10.1,1.0"},
			wantBuckets: []float64{1.0, 5.0, 6.9},
		},
		{
			name:        "well, empty bucket list return a empty slice",
			args:        args{""},
			wantBuckets: []float64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if gotBuckets := parseBuckets(tt.args.s); !reflect.DeepEqual(gotBuckets, tt.wantBuckets) {
				t.Errorf("parseBuckets() = %v, want %v", gotBuckets, tt.wantBuckets)
			}
		})
	}
}

func Test_parseObjectives(t *testing.T) {
	type args struct {
		s string
	}
	tests := []struct {
		name string
		args args
		want map[float64]float64
	}{
		{
			name: "objectives can be parsed",
			args: args{"0.5:0.05,0.9:0.01,0.99:0.001"},
			want: map[float64]float64{
				0.5:  0.05,
				0.9:  0.01,
				0.99: 0.001,
			},
		},
		{
			name: "invalid objectives in list will be ignores",
			args: args{"0.5:0.05,0.9:0.01,seems:invalid,invalid:0.6,0.9:also-invalid"},
			want: map[float64]float64{
				0.5: 0.05,
				0.9: 0.01,
			},
		},
		{
			name: "objectives are empty when nothing is given",
			args: args{""},
			want: map[float64]float64{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseObjectives(tt.args.s); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseObjectives() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_parseStepWidth(t *testing.T) {

	requestWithStepWidth, _ := http.NewRequest("PUT", "foo.de/test?add=3.4", strings.NewReader(""))
	requestWithoutStepWidth, _ := http.NewRequest("PUT", "foo.de/test?without=1", strings.NewReader(""))
	requestWithInvalidStepWidth, _ := http.NewRequest("PUT", "foo.de/test?add=beer", strings.NewReader(""))

	type args struct {
		request *http.Request
	}
	tests := []struct {
		name string
		args args
		want float64
	}{
		{
			name: "step width can be parsed",
			args: args{request: requestWithStepWidth},
			want: 3.4,
		},
		{
			name: "assume 1.0 if no step width is given",
			args: args{requestWithoutStepWidth},
			want: 1,
		},
		{
			name: "assume 1.0 if step width has invalid value",
			args: args{requestWithInvalidStepWidth},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := parseStepWidth(tt.args.request); got != tt.want {
				t.Errorf("parseStepWidth() = %v, want %v", got, tt.want)
			}
		})
	}
}
