package services

import (
	"fmt"
	"io"
	"reflect"
	"testing"
	"time"

	"github.com/itaraxa/effectivepancake/internal/models"
)

func TestParseQueryString(t *testing.T) {
	type args struct {
		raw string
	}
	tests := []struct {
		name    string
		args    args
		wantQ   Querier
		wantErr bool
	}{
		// TODO: Add test cases.
		{
			name: `Update gauge`,
			args: args{raw: `/update/gauge/test1/3.14`},
			wantQ: &models.Query{
				// raw:            `/update/gauge/test1/3.14`,
				MetricType:     `gauge`,
				MetricName:     `test1`,
				MetricRawValue: `3.14`,
			},
			wantErr: false,
		},
		{
			name: `Update counter`,
			args: args{raw: `/update/counter/test2/3`},
			wantQ: &models.Query{
				// raw:            `/update/gauge/test2/3`,
				MetricType:     `counter`,
				MetricName:     `test2`,
				MetricRawValue: `3`,
			},
			wantErr: false,
		},
		{
			name:    `Gauge without value`,
			args:    args{raw: `/update/gauge/test4/`},
			wantQ:   &models.Query{},
			wantErr: true,
		},
		{
			name:    `Сounter without value`,
			args:    args{raw: `/update/counter/test3/`},
			wantQ:   &models.Query{},
			wantErr: true,
		},
		{
			name:    `Bad type`,
			args:    args{raw: `/update/badtype/test5/14`},
			wantQ:   &models.Query{},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotQ, err := ParseQueryString(tt.args.raw)

			if (err != nil) != tt.wantErr {
				fmt.Println(gotQ.String())
				t.Errorf("ParseQueryString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotQ, tt.wantQ) && err == nil {
				t.Errorf("ParseQueryString() = %v, want %v", gotQ, tt.wantQ)
			}
		})
	}
}

func TestLoadMetrics(t *testing.T) {
	type args struct {
		mu  MetricUpdater
		src io.Reader
	}
	tests := []struct {
		name    string
		args    args
		want    time.Time
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := LoadMetrics(tt.args.mu, tt.args.src)
			if (err != nil) != tt.wantErr {
				t.Errorf("LoadMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LoadMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLoadMetricsFromFile(t *testing.T) {
	type args struct {
		l        logger
		mu       MetricUpdater
		fileName string
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := LoadMetricsFromFile(tt.args.l, tt.args.mu, tt.args.fileName); (err != nil) != tt.wantErr {
				t.Errorf("LoadMetricsFromFile() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
