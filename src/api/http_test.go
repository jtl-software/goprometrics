package api

import (
	"goprometrics/src/store"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

func Test_createPrometheusMetricOpts(t *testing.T) {

	request, _ := http.NewRequest("PUT", "http://127.0.0.1/count/myns/myname", strings.NewReader(""))

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
		name      string
		args      args
		wantOpts  store.MetricOpts
		wantValue float64
		wantErr   bool
	}{
		{
			name: "can parse namespace and name",
			args: args{request, map[string]string{"ns": "myns", "name": "myname"}},
			wantOpts: store.MetricOpts{
				Ns:                "myns",
				Name:              "myname",
				Label:             store.ConstLabel{},
				Help:              "",
				HistogramBuckets:  []float64{},
				SummaryObjectives: map[float64]float64{},
			},
			wantValue: 1,
			wantErr:   false,
		},
		{
			name: "can parse from body",
			args: args{requestWithLabels, map[string]string{}},
			wantOpts: store.MetricOpts{
				Ns:   "",
				Name: "",
				Label: store.ConstLabel{
					Name:  []string{"beer", "foo"},
					Value: []string{"wine", "bar"},
				},
				Help:             "beer",
				HistogramBuckets: []float64{1, 2},
				SummaryObjectives: map[float64]float64{
					0.5:  0.6,
					0.99: 0.10,
				},
			},
			wantValue: 1,
			wantErr:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotOpts, gotValue, err := createPrometheusMetricOpts(tt.args.r, tt.args.v)
			if (err != nil) != tt.wantErr {
				t.Errorf("createPrometheusMetricOpts() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotOpts, tt.wantOpts) {
				t.Errorf("createPrometheusMetricOpts() gotOpts = %v, want %v", gotOpts, tt.wantOpts)
			}
			if gotValue != tt.wantValue {
				t.Errorf("createPrometheusMetricOpts() gotValue = %v, want %v", gotValue, tt.wantValue)
			}
		})
	}
}

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
