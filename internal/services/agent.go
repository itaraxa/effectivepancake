package services

import (
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"

	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

/*
Sending metrica data to server via http

Args:

	ms *models.Metrics: pointer to models.Metrics object, which store metrica data
	serverURL string: endpoint of server
	client *http.Client: pointer to httpClient object, which uses for connection to server

Returns:

	error: nil or error, encountered during sending data
*/
func SendMetricsToServer(ms *models.Metrics, serverURL string, client *http.Client) error {
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
func CollectMetrics(pollCount uint64) (*models.Metrics, error) {
	ms := &models.Metrics{}

	ms.AddPollCount(pollCount)
	ms.AddData(collectRuntimeMetrics())
	ms.AddData(collectOtherMetrics())

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