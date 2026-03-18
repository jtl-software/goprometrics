package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"

	"goprometrics/src/store"
)

func Test_createPrometheusMetricOpts(t *testing.T) {

	request, _ := http.NewRequest("PUT", "http://127.0.0.1/count/myns/myname", strings.NewReader(""))

	requestWithLabels, _ := http.NewRequest(
		"PUT",
		"foo.de",
		strings.NewReader("labels=foo:bar,beer:wine&objectives=0.5:0.6,0.99:0.10&buckets=1,2&help=beer"))
	requestWithLabels.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	requestWithLabelsContainColonInValues, _ := http.NewRequest(
		"PUT",
		"foo.de",
		strings.NewReader("labels=foo:bar,beer:wi:ne,wine:red&objectives=0.5:0.6,0.99:0.10&buckets=1,2&help=beer"))
	requestWithLabelsContainColonInValues.Header.Add("Content-Type", "application/x-www-form-urlencoded")

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
		{
			name: "can handle colon in label value correctly",
			args: args{requestWithLabelsContainColonInValues, map[string]string{}},
			wantOpts: store.MetricOpts{
				Ns:   "",
				Name: "",
				Label: store.ConstLabel{
					Name:  []string{"beer", "foo", "wine"},
					Value: []string{"wi:ne", "bar", "red"},
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

// ---------------------------------------------------------------------------
// Additional cases for Test_createPrometheusMetricOpts
// ---------------------------------------------------------------------------

func Test_createPrometheusMetricOpts_non_numeric_value_returns_error(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/observe/ns/name/notanumber", strings.NewReader(""))
	_, _, err := createPrometheusMetricOpts(req, map[string]string{"ns": "ns", "name": "name", "value": "notanumber"})
	if err == nil {
		t.Errorf("expected error for non-numeric path value, got nil")
	}
}

func Test_createPrometheusMetricOpts_useSet_sets_flag(t *testing.T) {
	req, _ := http.NewRequest("PUT", "/gauge/ns/name/5", strings.NewReader("useSet=1"))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	opts, _, err := createPrometheusMetricOpts(req, map[string]string{"ns": "ns", "name": "name", "value": "5"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !opts.SetGaugeToValue {
		t.Errorf("SetGaugeToValue should be true when useSet=1 is in the form body")
	}
}

// ---------------------------------------------------------------------------
// Test_createLabels
// ---------------------------------------------------------------------------

func Test_createLabels(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  store.ConstLabel
	}{
		{
			name:  "empty input returns empty label",
			input: "",
			want:  store.ConstLabel{},
		},
		{
			name:  "single label parsed correctly",
			input: "env:prod",
			want:  store.ConstLabel{Name: []string{"env"}, Value: []string{"prod"}},
		},
		{
			name:  "multiple labels are sorted by name",
			input: "status:200,path:/login",
			want: store.ConstLabel{
				Name:  []string{"path", "status"},
				Value: []string{"/login", "200"},
			},
		},
		{
			name:  "colon in value is preserved",
			input: "url:http://example.com",
			want: store.ConstLabel{
				Name:  []string{"url"},
				Value: []string{"http://example.com"},
			},
		},
		{
			name:  "entries without colon are skipped",
			input: "valid:yes,nocolon,also:ok",
			want: store.ConstLabel{
				Name:  []string{"also", "valid"},
				Value: []string{"ok", "yes"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := createLabels(tt.input)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("createLabels(%q) = %+v, want %+v", tt.input, got, tt.want)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test_handleResponse
// ---------------------------------------------------------------------------

func Test_handleResponse(t *testing.T) {
	tests := []struct {
		name     string
		created  bool
		wantCode int
	}{
		{name: "created=true writes 201", created: true, wantCode: http.StatusCreated},
		{name: "created=false writes 200", created: false, wantCode: http.StatusOK},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			handleResponse(tt.created, rec)
			if rec.Code != tt.wantCode {
				t.Errorf("handleResponse(%v) wrote status %d, want %d", tt.created, rec.Code, tt.wantCode)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// Test_handleBadRequestError
// ---------------------------------------------------------------------------

func Test_handleBadRequestError(t *testing.T) {
	rec := httptest.NewRecorder()
	handleBadRequestError(fmt.Errorf("something went wrong"), rec)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("want status 400, got %d", rec.Code)
	}

	var body struct {
		Message string `json:"message"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("response body is not valid JSON: %v", err)
	}
	if body.Message != "something went wrong" {
		t.Errorf("want message %q, got %q", "something went wrong", body.Message)
	}
}

// ---------------------------------------------------------------------------
// Additional cases for Test_parseStepWidth (<=0 guard)
// ---------------------------------------------------------------------------

func Test_parseStepWidth_zero_defaults_to_one(t *testing.T) {
	req, _ := http.NewRequest("PUT", "foo.de/test?add=0", strings.NewReader(""))
	if got := parseStepWidth(req); got != 1.0 {
		t.Errorf("parseStepWidth with add=0: want 1.0, got %f", got)
	}
}

func Test_parseStepWidth_negative_defaults_to_one(t *testing.T) {
	req, _ := http.NewRequest("PUT", "foo.de/test?add=-5", strings.NewReader(""))
	if got := parseStepWidth(req); got != 1.0 {
		t.Errorf("parseStepWidth with add=-5: want 1.0, got %f", got)
	}
}

// ---------------------------------------------------------------------------
// Additional case for Test_parseObjectives (no-colon input — was a panic)
// ---------------------------------------------------------------------------

func Test_parseObjectives_no_colon_is_skipped(t *testing.T) {
	// "0.5" alone has no colon — must return empty map, not panic.
	got := parseObjectives("0.5")
	if len(got) != 0 {
		t.Errorf("parseObjectives with no-colon input should return empty map, got %v", got)
	}
}
