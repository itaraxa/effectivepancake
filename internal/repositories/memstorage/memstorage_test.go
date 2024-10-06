package memstorage

import (
	"testing"
)

func TestMemStorage_UpdateGauge(t *testing.T) {
	type fields struct {
		Gauge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		metricName string
		value      float64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Add normal gauge value",
			fields: fields{
				Gauge: map[string]float64{
					`testGauge`: 3.14,
				},
				Counter: map[string]int64{},
			},
			args: args{
				metricName: `testGauge`,
				value:      3.14,
			},
			wantErr: false,
		},
		{
			name: "Add normal gauge value into existing metric",
			fields: fields{
				Gauge: map[string]float64{
					`testGauge`: 8.88,
				},
				Counter: map[string]int64{},
			},
			args: args{
				metricName: `testGauge`,
				value:      8.88,
			},
			wantErr: false,
		},
		{
			name: "Add zero value",
			fields: fields{
				Gauge: map[string]float64{
					`zeroGauge`: 3.14,
				},
				Counter: map[string]int64{},
			},
			args: args{
				metricName: `zeroGauge`,
				value:      0.0,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauge:   tt.fields.Gauge,
				Counter: tt.fields.Counter,
			}
			if err := m.UpdateGauge(tt.args.metricName, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.UpdateGauge() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_AddCounter(t *testing.T) {
	type fields struct {
		Gauge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		metricName string
		value      int64
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "Add normal counter value",
			fields: fields{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{`testCounter`: 1},
			},
			args: args{
				metricName: `testCounter`,
				value:      1,
			},
			wantErr: false,
		},
		{
			name: "Add second normal counter value",
			fields: fields{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{`testCounter`: 3},
			},
			args: args{
				metricName: `testCounter`,
				value:      2,
			},
			wantErr: false,
		},
		{
			name: "Add zero counter value",
			fields: fields{
				Gauge:   map[string]float64{},
				Counter: map[string]int64{`testCounter`: 3},
			},
			args: args{
				metricName: `testCounter`,
				value:      0,
			},
			wantErr: false,
		},
		{
			name: "Add new normal counter value",
			fields: fields{
				Gauge: map[string]float64{},
				Counter: map[string]int64{
					`testCounter`:  3,
					`testCounter2`: 1,
				},
			},
			args: args{
				metricName: `testCounter2`,
				value:      1,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauge:   tt.fields.Gauge,
				Counter: tt.fields.Counter,
			}
			if err := m.AddCounter(tt.args.metricName, tt.args.value); (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.AddCounter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestMemStorage_GetMetrica(t *testing.T) {
	type fields struct {
		Gauge   map[string]float64
		Counter map[string]int64
	}
	type args struct {
		metricaType string
		metricaName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    string
		wantErr bool
	}{
		{
			name:   `Get not existing metrica`,
			fields: fields{},
			args: args{
				metricaType: `gauge`,
				metricaName: `test`,
			},
			wantErr: true,
		},
		{
			name: `Get existing gauge metrica`,
			fields: fields{
				Gauge: map[string]float64{`test`: 3.14},
			},
			args: args{
				metricaType: `gauge`,
				metricaName: `test`,
			},
			want:    `3.14`,
			wantErr: false,
		},
		{
			name: `Get existing counter metrica`,
			fields: fields{
				Counter: map[string]int64{`test`: 42},
			},
			args: args{
				metricaType: `counter`,
				metricaName: `test`,
			},
			want:    `42`,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				Gauge:   tt.fields.Gauge,
				Counter: tt.fields.Counter,
			}
			got, err := m.GetMetrica(tt.args.metricaType, tt.args.metricaName)
			if (err != nil) != tt.wantErr {
				t.Errorf("MemStorage.GetMetrica() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("MemStorage.GetMetrica() = %v, want %v", got, tt.want)
			}
		})
	}
}
