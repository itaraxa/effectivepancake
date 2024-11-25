package services

import (
	"net/http"
	"reflect"
	"sync"
	"testing"

	"github.com/itaraxa/effectivepancake/internal/config"
	"github.com/itaraxa/effectivepancake/internal/models"
)

func Test_sendMetricsToServerQueryStr(t *testing.T) {
	type args struct {
		l         logger
		ms        MetricsGetter
		serverURL string
		client    *http.Client
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
			if err := sendMetricsToServerQueryStr(tt.args.l, tt.args.ms, tt.args.serverURL, tt.args.client); (err != nil) != tt.wantErr {
				t.Errorf("sendMetricsToServerQueryStr() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sendMetricaToServerJSON(t *testing.T) {
	type args struct {
		l         logger
		ms        MetricsGetter
		serverURL string
		client    *http.Client
		key       string
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
			if err := sendMetricaToServerJSON(tt.args.l, tt.args.ms, tt.args.serverURL, tt.args.client, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("sendMetricaToServerJSON() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_sendMetricaToServerJSONgzip(t *testing.T) {
	type args struct {
		l         logger
		ms        MetricsGetter
		serverURL string
		client    *http.Client
		key       string
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
			if err := sendMetricaToServerJSONgzip(tt.args.l, tt.args.ms, tt.args.serverURL, tt.args.client, tt.args.key); (err != nil) != tt.wantErr {
				t.Errorf("sendMetricaToServerJSONgzip() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func Test_collectMetrics(t *testing.T) {
	type args struct {
		pollCount int64
	}
	tests := []struct {
		name    string
		args    args
		want    MetricsAddGetter
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := collectMetrics(tt.args.pollCount)
			if (err != nil) != tt.wantErr {
				t.Errorf("collectMetrics() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_collectRuntimeMetrics(t *testing.T) {
	tests := []struct {
		name string
		want []models.JSONMetric
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectRuntimeMetrics(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectRuntimeMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_collectOtherMetrics(t *testing.T) {
	tests := []struct {
		name string
		want []models.JSONMetric
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := collectOtherMetrics(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("collectOtherMetrics() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestPollMetrics(t *testing.T) {
	type args struct {
		wg          *sync.WaitGroup
		controlChan chan bool
		dataChan    chan MetricsAddGetter
		l           logger
		config      *config.AgentConfig
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			PollMetrics(tt.args.wg, tt.args.controlChan, tt.args.dataChan, tt.args.l, tt.args.config)
		})
	}
}

func TestReportMetrics(t *testing.T) {
	type args struct {
		wg          *sync.WaitGroup
		controlChan chan bool
		dataChan    chan MetricsAddGetter
		l           logger
		config      *config.AgentConfig
		client      *http.Client
	}
	tests := []struct {
		name string
		args args
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ReportMetrics(tt.args.wg, tt.args.controlChan, tt.args.dataChan, tt.args.l, tt.args.config, tt.args.client)
		})
	}
}
