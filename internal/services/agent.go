package services

import (
	"fmt"
	"io"
	"log/slog"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/itaraxa/effectivepancake/internal/config"
	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

// Интерфейс для работы с метриками на агенте
type Metricer interface {
	AddData(data []models.Metric) error
	AddPollCount(pollCount uint64) error
	String() string
	GetData() []models.Metric
}

/*
Sending metrica data to server via http

Args:

	ms Metricer: pointer to object implemented Metricer interface
	serverURL string: endpoint of server
	client *http.Client: pointer to httpClient object, which uses for connection to server

Returns:

	error: nil or error, encountered during sending data
*/
func sendMetricsToServer(ms Metricer, serverURL string, client *http.Client) error {
	for _, m := range ms.GetData() {
		res, err := client.Post(fmt.Sprintf("http://%s/update/%s/%s/%s", serverURL, m.Type, m.Name, m.Value), "text/plain", nil)
		if err != nil {
			return errors.ErrSendingMetricsToServer
		}
		// Reading response body to the end to Close body and release the TCP-connection
		_, err = io.Copy(io.Discard, res.Body)
		res.Body.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

/*
Collecting metrics. This function agregates and executes all ways for collecting metrica

Args:

	pollCount uint64: count for writing to PollCount metrica

Returns:

	*models.Metrica: pointer to models.Metrics structure, which store metrica data on Agent
	error: nil
*/
func collectMetrics(pollCount uint64) (Metricer, error) {
	ms := &models.Metrics{}

	err := ms.AddPollCount(pollCount)
	if err != nil {
		return ms, errors.ErrAddPollCount
	}
	err = ms.AddData(collectRuntimeMetrics())
	if err != nil {
		return ms, errors.ErrAddData
	}
	err = ms.AddData(collectOtherMetrics())
	if err != nil {
		return ms, errors.ErrAddData
	}

	return ms, nil
}

/*
Collecting metrics from runtime package.
Collected metrics:
  - Alloc
  - BuckHashSys
  - Frees
  - GCCPUFraction
  - GCSys
  - HeapAlloc
  - HeapIdle
  - HeapInuse
  - HeapObjects
  - HeapReleased
  - HeapSys
  - LastGC
  - Lookups
  - MCacheInuse
  - MCacheSys
  - MSpanInuse
  - MSpanSys
  - Mallocs
  - NextGC
  - NumForcedGC
  - NumGC
  - OtherSys
  - PauseTotalNs
  - StackInuse
  - StackSys
  - Sys
  - TotalAlloc

Args:

	None

Returns:

	[]models.Metrica: slice of models.Metrica objects
*/
func collectRuntimeMetrics() []models.Metric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	out := []models.Metric{}
	out = append(out, models.Metric{Name: "Alloc", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Alloc)})
	out = append(out, models.Metric{Name: "BuckHashSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.BuckHashSys)})
	out = append(out, models.Metric{Name: "Frees", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Frees)})
	out = append(out, models.Metric{Name: "GCCPUFraction", Type: "gauge", Value: fmt.Sprintf("%f", memStats.GCCPUFraction)})
	out = append(out, models.Metric{Name: "GCSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.GCSys)})
	out = append(out, models.Metric{Name: "HeapAlloc", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapAlloc)})
	out = append(out, models.Metric{Name: "HeapIdle", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapIdle)})
	out = append(out, models.Metric{Name: "HeapInuse", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapInuse)})
	out = append(out, models.Metric{Name: "HeapObjects", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapObjects)})
	out = append(out, models.Metric{Name: "HeapReleased", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapReleased)})
	out = append(out, models.Metric{Name: "HeapSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapSys)})
	out = append(out, models.Metric{Name: "LastGC", Type: "gauge", Value: fmt.Sprintf("%d", memStats.LastGC)})
	out = append(out, models.Metric{Name: "Lookups", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Lookups)})
	out = append(out, models.Metric{Name: "MCacheInuse", Type: "gauge", Value: fmt.Sprintf("%d", memStats.MCacheInuse)})
	out = append(out, models.Metric{Name: "MCacheSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.MCacheSys)})
	out = append(out, models.Metric{Name: "MSpanInuse", Type: "gauge", Value: fmt.Sprintf("%d", memStats.MSpanInuse)})
	out = append(out, models.Metric{Name: "MSpanSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.MSpanSys)})
	out = append(out, models.Metric{Name: "Mallocs", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Mallocs)})
	out = append(out, models.Metric{Name: "NextGC", Type: "gauge", Value: fmt.Sprintf("%d", memStats.NextGC)})
	out = append(out, models.Metric{Name: "NumForcedGC", Type: "gauge", Value: fmt.Sprintf("%d", memStats.NumForcedGC)})
	out = append(out, models.Metric{Name: "NumGC", Type: "gauge", Value: fmt.Sprintf("%d", memStats.NumGC)})
	out = append(out, models.Metric{Name: "OtherSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.OtherSys)})
	out = append(out, models.Metric{Name: "PauseTotalNs", Type: "gauge", Value: fmt.Sprintf("%d", memStats.PauseTotalNs)})
	out = append(out, models.Metric{Name: "StackInuse", Type: "gauge", Value: fmt.Sprintf("%d", memStats.StackInuse)})
	out = append(out, models.Metric{Name: "StackSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.StackSys)})
	out = append(out, models.Metric{Name: "Sys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Sys)})
	out = append(out, models.Metric{Name: "TotalAlloc", Type: "gauge", Value: fmt.Sprintf("%d", memStats.TotalAlloc)})

	return out
}

/*
Collecting other metrics.
Other metrics:
  - random value

Args:

	None

Returns:

	[]models.Metrica: slice of models.Metrica objects
*/
func collectOtherMetrics() []models.Metric {
	out := []models.Metric{}

	out = append(out, models.Metric{
		Name:  "RandomValue",
		Type:  "gauge",
		Value: fmt.Sprintf("%f", rand.Float64()),
	})

	return out
}

/*
Function for periodically collecting all metrics

Args:

	wg *sync.WaitGroup: pointer to sync.WaitGroup for for controlling the completion of a function in a goroutine
	controlChan chan bool: channel for receiving a stop signal
	dataChan chan Metricer: channel for exchanging metric data
	l *slog.Logger: pointer to logger instance
	pollInterval time.Duration

Returns:

	None
*/
func PollMetrics(wg *sync.WaitGroup, controlChan chan bool, dataChan chan Metricer, l *slog.Logger, config *config.AgentConfig) {
	defer wg.Done()
	var pollCounter uint64 = 0
POLLING:
	for {
		controlChan <- false

		l.Debug("Poll counter", slog.Uint64("Value", pollCounter))
		ms, err := collectMetrics(pollCounter)
		if err != nil {
			l.Error("Error collect metrics")
		}
		if len(dataChan) == cap(dataChan) {
			l.Error("Error internal commnication", slog.String("error", errors.ErrChannelFull.Error()))
		}
		dataChan <- ms
		pollCounter += 1
		time.Sleep(config.PollInterval)

		if <-controlChan {
			l.Info("Polling metrica stopped")
			break POLLING
		}
	}
}

/*
Function for periodically sending metrics

Args:

	wg *sync.WaitGroup: pointer to sync.WaitGroup for for controlling the completion of a function in a goroutine
	controlChan chan bool: channel for receiving a stop signal
	dataChan chan Metricer: channel for exchanging metric data
	l *slog.Logger: pointer to logger instance
	config *config.AgentConfig: pointer to config instance
	client *http.Client: pointer to http client instance

Returns:

	None
*/
func ReportMetrics(wg *sync.WaitGroup, controlChan chan bool, dataChan chan Metricer, l *slog.Logger, config *config.AgentConfig, client *http.Client) {
	defer wg.Done()
	var reportCounter uint64 = 0
REPORTING:
	for {
		controlChan <- false

		time.Sleep(config.ReportInterval)
		for len(dataChan) > 0 {
			l.Debug("Report counter", slog.Uint64("Value", reportCounter))
			err := sendMetricsToServer(<-dataChan, config.AddressServer, client)
			if err != nil {
				l.Error("Error sending to server. Waiting 1 second",
					slog.String("server", config.AddressServer),
					slog.String("error", errors.ErrSendingMetricsToServer.Error()),
				)
				// Pause for next sending
				time.Sleep(1 * time.Second)
			}
		}
		reportCounter++

		if <-controlChan {
			l.Info("Reporting metrica stopped")
			break REPORTING
		}
	}
}
