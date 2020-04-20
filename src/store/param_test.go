package store

import "testing"

func TestMetricOpts_Key(t *testing.T) {
	type fields struct {
		Ns    string
		Name  string
		Label ConstLabel
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: "Can build Key by Name",
			fields: fields{
				Ns:    "My",
				Name:  "Test",
				Label: ConstLabel{},
			},
			want: "My_Test__",
		},
		{
			name: "Can build Key by Name and Labels",
			fields: fields{
				Ns:   "My",
				Name: "Test",
				Label: ConstLabel{
					Name:  []string{"foo", "bar"},
					Value: []string{"1", "2"},
				},
			},
			want: "My_Test__foo_bar",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opts := MetricOpts{
				Ns:    tt.fields.Ns,
				Name:  tt.fields.Name,
				Label: tt.fields.Label,
			}
			if got := opts.Key(); got != tt.want {
				t.Errorf("Key() = %v, want %v", got, tt.want)
			}
		})
	}
}
