package models

import (
	"testing"
)

func TestQuery_String(t *testing.T) {
	type fields struct {
		raw            string
		MetricType     string
		MetricName     string
		MetricRawValue string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: `Test string output`,
			fields: fields{
				raw:            `update/gauge/test1/3.14`,
				MetricType:     `gauge`,
				MetricName:     `test1`,
				MetricRawValue: `3.14`,
			},
			want: "==== Query ====\n\r>Raw: update/gauge/test1/3.14\n\r>>Type: gauge\n\r>>Name: test1\n\r>>Value: 3.14\n\r",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := Query{
				raw:            tt.fields.raw,
				MetricType:     tt.fields.MetricType,
				MetricName:     tt.fields.MetricName,
				MetricRawValue: tt.fields.MetricRawValue,
			}
			if got := q.String(); got != tt.want {
				t.Errorf("Query.String() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuery_GetMetricaType(t *testing.T) {
	type fields struct {
		raw            string
		MetricType     string
		MetricName     string
		MetricRawValue string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: `Get gauge type`,
			fields: fields{
				MetricType: `gauge`,
			},
			want: `gauge`,
		},
		{
			name: `Get counter type`,
			fields: fields{
				MetricType: `counter`,
			},
			want: `counter`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := Query{
				raw:            tt.fields.raw,
				MetricType:     tt.fields.MetricType,
				MetricName:     tt.fields.MetricName,
				MetricRawValue: tt.fields.MetricRawValue,
			}
			if got := q.GetMetricaType(); got != tt.want {
				t.Errorf("Query.GetMetricaType() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuery_GetMetricName(t *testing.T) {
	type fields struct {
		raw            string
		MetricType     string
		MetricName     string
		MetricRawValue string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: `Get correct name 1`,
			fields: fields{
				MetricName: `test1`,
			},
			want: `test1`,
		},
		{
			name: `Get correct name 2`,
			fields: fields{
				MetricName: `test2`,
			},
			want: `test2`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := Query{
				raw:            tt.fields.raw,
				MetricType:     tt.fields.MetricType,
				MetricName:     tt.fields.MetricName,
				MetricRawValue: tt.fields.MetricRawValue,
			}
			if got := q.GetMetricName(); got != tt.want {
				t.Errorf("Query.GetMetricName() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQuery_GetMetricaRawValue(t *testing.T) {
	type fields struct {
		raw            string
		MetricType     string
		MetricName     string
		MetricRawValue string
	}
	tests := []struct {
		name   string
		fields fields
		want   string
	}{
		{
			name: `Get value 1`,
			fields: fields{
				MetricRawValue: `3.14`,
			},
			want: `3.14`,
		},
		{
			name: `Get value 2`,
			fields: fields{
				MetricRawValue: `42`,
			},
			want: `42`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			q := Query{
				raw:            tt.fields.raw,
				MetricType:     tt.fields.MetricType,
				MetricName:     tt.fields.MetricName,
				MetricRawValue: tt.fields.MetricRawValue,
			}
			if got := q.GetMetricaRawValue(); got != tt.want {
				t.Errorf("Query.GetMetricaRawValue() = %v, want %v", got, tt.want)
			}
		})
	}
}
