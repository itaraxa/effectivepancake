package services

import (
	"reflect"
	"testing"

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
				t.Errorf("ParseQueryString() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotQ, tt.wantQ) {
				t.Errorf("ParseQueryString() = %v, want %v", gotQ, tt.wantQ)
			}
		})
	}
}
