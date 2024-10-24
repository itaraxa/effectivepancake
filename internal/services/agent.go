package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/itaraxa/effectivepancake/internal/config"
	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/logger"
	"github.com/itaraxa/effectivepancake/internal/models"
)

// Интерфейс для работы с метриками на агенте
type Metricer interface {
	AddData(data []models.JSONMetric) error
	AddPollCount(pollCount int64) error
	String() string
	GetData() []models.JSONMetric
}

/*
sendMetricsToServerQueryStr send metrica data to server via http. Data included into request string

Args:

	ms Metricer: pointer to object implemented Metricer interface
	serverURL string: endpoint of server
	client *http.Client: pointer to httpClient object, which uses for connection to server

Returns:

	error: nil or error, encountered during sending data
*/
func sendMetricsToServerQueryStr(ms Metricer, serverURL string, client *http.Client) error {
	for _, m := range ms.GetData() {
		queryString := ""
		if m.MType == "gauge" {
			queryString = fmt.Sprintf("http://%s/update/%s/%s/%f", serverURL, m.MType, m.ID, m.Value)
		} else if m.MType == "counter" {
			queryString = fmt.Sprintf("http://%s/update/%s/%s/%d", serverURL, m.MType, m.ID, m.Delta)
		}
		res, err := client.Post(queryString, "text/plain", nil)
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
sendMetricaToServerJSON send metrica data to server via http POST method. Data included into request body in JSON

Args:

	ms Metricer: pointer to object implemented Metricer interface
	serverURL string: endpoint of server
	client *http.Client: pointer to httpClient object, which uses for connection to server

Returns:

	error: nil or error, encountered during sending data
*/
func sendMetricaToServerJSON(l logger.Logger, ms Metricer, serverURL string, client *http.Client) error {
	for _, m := range ms.GetData() {
		if m.Delta == 0 && m.Value == 0.0 {
			continue
		}

		jsondata, err := json.Marshal(m)
		if err != nil {
			return err
		}

		l.Debug("json data for send", "string representation", string(jsondata))

		resp, err := client.Post(fmt.Sprintf("http://%s/update/", serverURL), "application/json", bytes.NewBuffer(jsondata))
		if err != nil {
			return errors.ErrSendingMetricsToServer
		}
		_, err = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
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
func collectMetrics(pollCount int64) (Metricer, error) {
	jms := &models.JSONMetrics{}

	err := jms.AddPollCount(pollCount)
	if err != nil {
		return jms, errors.ErrAddPollCount
	}
	err = jms.AddData(collectRuntimeMetrics())
	if err != nil {
		return jms, errors.ErrAddData
	}
	err = jms.AddData(collectOtherMetrics())
	if err != nil {
		return jms, errors.ErrAddData
	}

	return jms, nil
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
func collectRuntimeMetrics() []models.JSONMetric {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	out := []models.JSONMetric{}
	out = append(out, models.JSONMetric{ID: "Alloc", MType: "gauge", Value: float64(memStats.Alloc)})
	out = append(out, models.JSONMetric{ID: "BuckHashSys", MType: "gauge", Value: float64(memStats.BuckHashSys)})
	out = append(out, models.JSONMetric{ID: "Frees", MType: "gauge", Value: float64(memStats.Frees)})
	out = append(out, models.JSONMetric{ID: "GCCPUFraction", MType: "gauge", Value: float64(memStats.GCCPUFraction)})
	out = append(out, models.JSONMetric{ID: "GCSys", MType: "gauge", Value: float64(memStats.GCSys)})
	out = append(out, models.JSONMetric{ID: "HeapAlloc", MType: "gauge", Value: float64(memStats.HeapAlloc)})
	out = append(out, models.JSONMetric{ID: "HeapIdle", MType: "gauge", Value: float64(memStats.HeapIdle)})
	out = append(out, models.JSONMetric{ID: "HeapInuse", MType: "gauge", Value: float64(memStats.HeapInuse)})
	out = append(out, models.JSONMetric{ID: "HeapObjects", MType: "gauge", Value: float64(memStats.HeapObjects)})
	out = append(out, models.JSONMetric{ID: "HeapReleased", MType: "gauge", Value: float64(memStats.HeapReleased)})
	out = append(out, models.JSONMetric{ID: "HeapSys", MType: "gauge", Value: float64(memStats.HeapSys)})
	out = append(out, models.JSONMetric{ID: "LastGC", MType: "gauge", Value: float64(memStats.LastGC)})
	out = append(out, models.JSONMetric{ID: "Lookups", MType: "gauge", Value: float64(memStats.Lookups)})
	out = append(out, models.JSONMetric{ID: "MCacheInuse", MType: "gauge", Value: float64(memStats.MCacheInuse)})
	out = append(out, models.JSONMetric{ID: "MCacheSys", MType: "gauge", Value: float64(memStats.MCacheSys)})
	out = append(out, models.JSONMetric{ID: "MSpanInuse", MType: "gauge", Value: float64(memStats.MSpanInuse)})
	out = append(out, models.JSONMetric{ID: "MSpanSys", MType: "gauge", Value: float64(memStats.MSpanSys)})
	out = append(out, models.JSONMetric{ID: "Mallocs", MType: "gauge", Value: float64(memStats.Mallocs)})
	out = append(out, models.JSONMetric{ID: "NextGC", MType: "gauge", Value: float64(memStats.NextGC)})
	out = append(out, models.JSONMetric{ID: "NumForcedGC", MType: "gauge", Value: float64(memStats.NumForcedGC)})
	out = append(out, models.JSONMetric{ID: "NumGC", MType: "gauge", Value: float64(memStats.NumGC)})
	out = append(out, models.JSONMetric{ID: "OtherSys", MType: "gauge", Value: float64(memStats.OtherSys)})
	out = append(out, models.JSONMetric{ID: "PauseTotalNs", MType: "gauge", Value: float64(memStats.PauseTotalNs)})
	out = append(out, models.JSONMetric{ID: "StackInuse", MType: "gauge", Value: float64(memStats.StackInuse)})
	out = append(out, models.JSONMetric{ID: "StackSys", MType: "gauge", Value: float64(memStats.StackSys)})
	out = append(out, models.JSONMetric{ID: "Sys", MType: "gauge", Value: float64(memStats.Sys)})
	out = append(out, models.JSONMetric{ID: "TotalAlloc", MType: "gauge", Value: float64(memStats.TotalAlloc)})

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
func collectOtherMetrics() []models.JSONMetric {
	out := []models.JSONMetric{}

	out = append(out, models.JSONMetric{
		ID:    "RandomValue",
		MType: "gauge",
		Value: rand.Float64(),
	})

	return out
}

/*
Function for periodically collecting all metrics

Args:

	wg *sync.WaitGroup: pointer to sync.WaitGroup for for controlling the completion of a function in a goroutine
	controlChan chan bool: channel for receiving a stop signal
	dataChan chan Metricer: channel for exchanging metric data
	l logger.Logger: pointer to logger instance
	pollInterval time.Duration

Returns:

	None
*/
func PollMetrics(wg *sync.WaitGroup, controlChan chan bool, dataChan chan Metricer, l logger.Logger, config *config.AgentConfig) {
	defer wg.Done()
	var pollCounter int64 = 0
POLLING:
	for {
		controlChan <- false

		l.Debug("Poll counter", "Value", pollCounter)
		ms, err := collectMetrics(pollCounter)
		if err != nil {
			l.Error("Error collect metrics")
		}
		if len(dataChan) == cap(dataChan) {
			l.Error("Error internal commnication", "error", errors.ErrChannelFull.Error())
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
	l logger.Logger: pointer to logger instance
	config *config.AgentConfig: pointer to config instance
	client *http.Client: pointer to http client instance

Returns:

	None
*/
func ReportMetrics(wg *sync.WaitGroup, controlChan chan bool, dataChan chan Metricer, l logger.Logger, config *config.AgentConfig, client *http.Client) {
	defer wg.Done()
	var reportCounter uint64 = 0
REPORTING:
	for {
		controlChan <- false

		time.Sleep(config.ReportInterval)
		for len(dataChan) > 0 {
			l.Debug("Report counter", "Value", reportCounter)
			switch config.ReportMode {
			case `json`:
				err := sendMetricaToServerJSON(l, <-dataChan, config.AddressServer, client)
				if err != nil {
					l.Error("Error sending to server. Waiting 1 second", "server", config.AddressServer, "error", errors.ErrSendingMetricsToServer.Error())
				}
			case `raw`:
				err := sendMetricsToServerQueryStr(<-dataChan, config.AddressServer, client)
				if err != nil {
					l.Error("Error sending to server. Waiting 1 second", "server", config.AddressServer, "error", errors.ErrSendingMetricsToServer.Error())
				}
			}
			// Pause for next sending
			time.Sleep(1 * time.Second)
		}
		reportCounter++

		if <-controlChan {
			l.Info("Reporting metrica stopped")
			break REPORTING
		}
	}
}
