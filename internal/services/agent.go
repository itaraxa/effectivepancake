package services

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/v4/cpu"
	"github.com/shirou/gopsutil/v4/mem"

	"github.com/itaraxa/effectivepancake/internal/config"
	myErrors "github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

/*
sendMetricsToServerQueryStr send metrica data to server via http. Data included into request string

Args:

	ms MetricsGetter: pointer to object implemented MetricsGetter interface
	serverURL string: endpoint of server
	client *http.Client: pointer to httpClient object, which uses for connection to server

Returns:

	error: nil or error, encountered during sending data
*/
func sendMetricsToServerQueryStr(l logger, ms MetricsGetter, serverURL string, client *http.Client) error {
	mData := ms.GetData()
	if len(mData) == 0 {
		return myErrors.ErrNoMetrics
	}
	for _, m := range mData {
		queryString := ""
		if m.MType == gauge {
			queryString = createURL(serverURL, m.MType, m.ID, fmt.Sprint(*m.Value))
		} else if m.MType == counter {
			queryString = createURL(serverURL, m.MType, m.ID, fmt.Sprint(*m.Delta))
		}
		l.Debug("query string", "string", queryString)
		req, err := http.NewRequest(`POST`, queryString, nil)
		req.Header.Set("Content-Type", "text/plain")
		if err != nil {
			l.Error("cannot create request", "error", err.Error())
			return err
		}
		resp, err := retryRequest(func() (*http.Response, error) { return client.Do(req) })
		if err != nil {
			return errors.Join(myErrors.ErrSendingMetricsToServer, err)
		}
		// Reading response body to the end to Close body and release the TCP-connection
		_, err = io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
		if err != nil {
			return err
		}
	}

	return nil
}

/*
sendMetricaToServerJSON send metrica data to server via http POST method. Data included into request body in JSON

Args:

	l logger: implementation of logger interface
	ms MetricsGetter: pointer to object implemented MetricsGetter interface
	serverURL string: endpoint of server
	client *http.Client: pointer to httpClient object, which uses for connection to server

Returns:

	error: nil or error, encountered during sending data
*/
func sendMetricaToServerJSON(l logger, ms MetricsGetter, serverURL string, client *http.Client, key string) error {
	mData := ms.GetData()
	if len(mData) == 0 {
		return myErrors.ErrNoMetrics
	}
	for _, m := range mData {
		jsonDataReq, err := json.Marshal(m)
		if err != nil {
			return err
		}

		l.Info("json data for send", "string representation", string(jsonDataReq))
		req, err := http.NewRequest(`POST`, createURL(serverURL, `update/`), bytes.NewBuffer(jsonDataReq))
		l.Debug("query string", "string", createURL(serverURL, `update/`))
		req.Header.Set("Content-Type", "application/json")
		if err != nil {
			l.Error("cannot create request", "error", err.Error())
			return err
		}
		if key != `` {
			signature, err := SignSHA256(l, bytes.NewReader(jsonDataReq), key)
			if err != nil {
				l.Error("cannot sign data", "error", err.Error())
				return err
			}
			req.Header.Set(`HashSHA256`, signature)
		}
		resp, err := retryRequest(func() (*http.Response, error) { return client.Do(req) })
		if err != nil {
			return errors.Join(myErrors.ErrSendingMetricsToServer, err)
		}
		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			l.Error("cannot read responce body")
			return err
		}
		l.Info("json data from responce", "string representation", buf.String())
		resp.Body.Close()
	}
	return nil
}

func sendMetricaToServerBatch(l logger, ms MetricsGetter, serverURL string, client *http.Client, key string) error {
	mData := ms.GetData()
	if len(mData) == 0 {
		return myErrors.ErrNoMetrics
	}
	jsonDataReq, err := json.Marshal(mData)
	if err != nil {
		return err
	}
	l.Debug("json data for send", "string representation", string(jsonDataReq))
	req, err := http.NewRequest(`POST`, createURL(serverURL, `updates/`), bytes.NewBuffer(jsonDataReq))
	l.Debug("query string", "string", createURL(serverURL, `updates/`))
	req.Header.Set("Content-Type", "application/json")
	if err != nil {
		l.Error("cannot create request", "error", err.Error())
		return err
	}
	if key != `` {
		signature, err := SignSHA256(l, bytes.NewReader(jsonDataReq), key)
		if err != nil {
			l.Error("cannot sign data", "error", err.Error())
			return err
		}
		req.Header.Set(`HashSHA256`, signature)
	}
	resp, err := retryRequest(func() (*http.Response, error) { return client.Do(req) })
	if err != nil {
		return errors.Join(myErrors.ErrSendingMetricsToServer, err)
	}
	var buf bytes.Buffer
	_, err = buf.ReadFrom(resp.Body)
	if err != nil {
		l.Error("cannot read responce body")
		return err
	}
	l.Info("json data from responce", "string representation", buf.String())
	resp.Body.Close()

	return nil
}

/*
sendMetricaToServerJSONgzip send metrica data to server via http POST method. Data included into request body in compressed JSON.
The response is checked for compression, and based on the result, it is decoded accordingly.

Args:

	l logger: implementation of logger-interface
	ms MetricsGetter: pointer to object implemented MetricsGetter interface
	serverURL string: endpoint of server
	client *http.Client: pointer to httpClient object, which uses for connection to server

Returns:

	error: nil or error, encountered during sending data
*/
func sendMetricaToServerJSONgzip(l logger, ms MetricsGetter, serverURL string, client *http.Client, key string) error {
	mData := ms.GetData()
	if len(mData) == 0 {
		return myErrors.ErrNoMetrics
	}
	for _, m := range mData {
		jsonDataReq, err := json.Marshal(m)
		if err != nil {
			return err
		}

		jsonGzipDataReq, err := compress(jsonDataReq)
		if err != nil {
			l.Error("cannot compress data", "error", err.Error())
		}
		l.Info("json data for send compressd", "string representation", string(jsonDataReq), "compress ratio", float64(len(jsonDataReq))/float64(len(jsonGzipDataReq)))
		req, err := http.NewRequest(`POST`, createURL(serverURL, `update/`), bytes.NewBuffer(jsonGzipDataReq))
		l.Debug("query string", "string", createURL(serverURL, `update/`))
		req.Header.Set("Accept-Encoding", "gzip")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Content-Encoding", "gzip")
		if err != nil {
			l.Error("cannot create request", "error", err.Error())
			return err
		}
		if key != `` {
			signature, err := SignSHA256(l, bytes.NewReader(jsonGzipDataReq), key)
			if err != nil {
				l.Error("cannot sign data", "error", err.Error())
				return err
			}
			req.Header.Set(`HashSHA256`, signature)
		}
		start := time.Now()
		resp, err := retryRequest(func() (*http.Response, error) { return client.Do(req) })
		if err != nil {
			return errors.Join(myErrors.ErrSendingMetricsToServer, err)
		}
		defer resp.Body.Close()

		switch {
		// get compressed data from server
		case resp.StatusCode == http.StatusOK && resp.Header.Get("Content-Encoding") == `gzip`:
			var buf bytes.Buffer
			_, err = buf.ReadFrom(resp.Body)
			if err != nil {
				l.Error("cannot read responce body")
				return myErrors.ErrGettingAnswerFromServer
			}
			data, err := decompress(buf.Bytes())
			if err != nil {
				l.Error("cannot decompress responce body")
				return myErrors.ErrGettingAnswerFromServer
			}
			l.Info("json data from responce", "string representation", string(data), "duration", time.Since(start))
		// get uncompressed data from server
		case resp.StatusCode == http.StatusOK && resp.Header.Get("Content-Encoding") == "":
			var buf bytes.Buffer
			_, err = buf.ReadFrom(resp.Body)
			if err != nil {
				l.Error("cannot read responce body")
				return myErrors.ErrGettingAnswerFromServer
			}
			l.Info("json data from responce", "string representation", buf.String(), "duration", time.Since(start))
		default:
			l.Info("received a response with an error code", "status code", resp.StatusCode, "duration", time.Since(start))
		}
	}
	return nil
}

func sendMetricaToServerBatchgzip(l logger, ms MetricsGetter, serverURL string, client *http.Client, key string) error {
	mData := ms.GetData()
	if len(mData) == 0 {
		l.Error("no metrics for sending")
		return myErrors.ErrNoMetrics
	}
	jsonDataReq, err := json.Marshal(mData)
	if err != nil {
		l.Error("marshalling data error", "error", err.Error())
		return err
	}
	jsonGzipDataReq, err := compress(jsonDataReq)
	if err != nil {
		l.Error("cannot compress data", "error", err.Error())
		return err
	}
	l.Debug("json data for send compressd", "string representation", string(jsonDataReq), "compress ratio", float64(len(jsonDataReq))/float64(len(jsonGzipDataReq)))

	req, err := http.NewRequest(`POST`, createURL(serverURL, `updates/`), bytes.NewBuffer(jsonGzipDataReq))
	l.Debug("query string", "string", createURL(serverURL, `updates/`))
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Content-Encoding", "gzip")
	if key != `` {
		signature, err := SignSHA256(l, bytes.NewReader(jsonGzipDataReq), key)
		if err != nil {
			l.Error("cannot sign data", "error", err.Error())
			return err
		}
		req.Header.Set(`HashSHA256`, signature)
	}
	if err != nil {
		l.Error("cannot create request", "error", err.Error())
		return err
	}

	start := time.Now()
	resp, err := retryRequest(func() (*http.Response, error) { return client.Do(req) })
	if err != nil {
		return errors.Join(myErrors.ErrSendingMetricsToServer, err)
	}
	defer resp.Body.Close()

	switch {
	// get compressed data from server
	case resp.StatusCode == http.StatusOK && resp.Header.Get("Content-Encoding") == `gzip`:
		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			l.Error("cannot read responce body")
			return myErrors.ErrGettingAnswerFromServer
		}
		data, err := decompress(buf.Bytes())
		if err != nil {
			l.Error("cannot decompress responce body")
			return myErrors.ErrGettingAnswerFromServer
		}
		l.Info("json data from responce", "string representation", string(data), "duration", time.Since(start))
	// get uncompressed data from server
	case resp.StatusCode == http.StatusOK && resp.Header.Get("Content-Encoding") == "":
		var buf bytes.Buffer
		_, err = buf.ReadFrom(resp.Body)
		if err != nil {
			l.Error("cannot read responce body")
			return myErrors.ErrGettingAnswerFromServer
		}
		l.Info("json data from responce", "string representation", buf.String(), "duration", time.Since(start))
	default:
		l.Info("received a response with an error code", "status code", resp.StatusCode, "duration", time.Since(start))
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
func collectMetrics(l logger, pollCount int64) (MetricsAddGetter, error) {
	jms := &models.JSONMetrics{}

	l.Debug("adding poll count")
	err := jms.AddPollCount(pollCount)
	if err != nil {
		return jms, myErrors.ErrAddPollCount
	}
	l.Debug("collecting runtime metrics")
	err = jms.AddData(collectRuntimeMetrics())
	if err != nil {
		return jms, myErrors.ErrAddData
	}
	l.Debug("colecting other metrics")
	err = jms.AddData(collectOtherMetrics())
	if err != nil {
		return jms, myErrors.ErrAddData
	}
	l.Debug("collecting GoPsUtils metrics")
	t, err := collectGoPsUtilMetrics(l)
	if err != nil {
		return jms, errors.Join(myErrors.ErrAddData, err)
	}
	err = jms.AddData(t)
	if err != nil {
		return jms, myErrors.ErrAddData
	}
	l.Debug("metrics collected")

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

	Alloc := float64(memStats.Alloc)
	BuckHashSys := float64(memStats.BuckHashSys)
	Frees := float64(memStats.Frees)
	GCCPUFraction := float64(memStats.GCCPUFraction)
	GCSys := float64(memStats.GCSys)
	HeapAlloc := float64(memStats.HeapAlloc)
	HeapIdle := float64(memStats.HeapIdle)
	HeapInuse := float64(memStats.HeapInuse)
	HeapObjects := float64(memStats.HeapObjects)
	HeapReleased := float64(memStats.HeapReleased)
	HeapSys := float64(memStats.HeapSys)
	LastGC := float64(memStats.LastGC)
	Lookups := float64(memStats.Lookups)
	MCacheInuse := float64(memStats.MCacheInuse)
	MCacheSys := float64(memStats.MCacheSys)
	MSpanInuse := float64(memStats.MSpanInuse)
	MSpanSys := float64(memStats.MSpanSys)
	Mallocs := float64(memStats.Mallocs)
	NextGC := float64(memStats.NextGC)
	NumForcedGC := float64(memStats.NumForcedGC)
	NumGC := float64(memStats.NumGC)
	OtherSys := float64(memStats.OtherSys)
	PauseTotalNs := float64(memStats.PauseTotalNs)
	StackInuse := float64(memStats.StackInuse)
	StackSys := float64(memStats.StackSys)
	Sys := float64(memStats.Sys)
	TotalAlloc := float64(memStats.TotalAlloc)

	metrics := []struct {
		ID    string
		Value *float64
	}{
		{"Alloc", &Alloc},
		{"BuckHashSys", &BuckHashSys},
		{"Frees", &Frees},
		{"GCCPUFraction", &GCCPUFraction},
		{"GCSys", &GCSys},
		{"HeapAlloc", &HeapAlloc},
		{"HeapIdle", &HeapIdle},
		{"HeapInuse", &HeapInuse},
		{"HeapObjects", &HeapObjects},
		{"HeapReleased", &HeapReleased},
		{"HeapSys", &HeapSys},
		{"LastGC", &LastGC},
		{"Lookups", &Lookups},
		{"MCacheInuse", &MCacheInuse},
		{"MCacheSys", &MCacheSys},
		{"MSpanInuse", &MSpanInuse},
		{"MSpanSys", &MSpanSys},
		{"Mallocs", &Mallocs},
		{"NextGC", &NextGC},
		{"NumForcedGC", &NumForcedGC},
		{"NumGC", &NumGC},
		{"OtherSys", &OtherSys},
		{"PauseTotalNs", &PauseTotalNs},
		{"StackInuse", &StackInuse},
		{"StackSys", &StackSys},
		{"Sys", &Sys},
		{"TotalAlloc", &TotalAlloc},
	}

	out := []models.JSONMetric{}
	for _, metric := range metrics {
		out = append(out, models.JSONMetric{
			ID:    metric.ID,
			MType: gauge,
			Value: metric.Value,
		})
	}

	return out
}

/*
collectGoPsUtilMetrics collects metrics from github.com/shirou/gopsutil/v4 package.
Collected metrics:
- TotalMemory
- FreeMemory
- CPUutilization#

Args:

	None

Returns:

	[]models.Metrica: slice of models.Metrica objects
	error
*/
func collectGoPsUtilMetrics(l logger) (out []models.JSONMetric, errs error) {
	l.Info("getting virtual memory info")
	vMemory, err := mem.VirtualMemory()
	if err != nil {
		l.Error("getting virtual memory info", "error", err.Error())
		errs = errors.Join(errs, myErrors.ErrGetVirtualMemory, err)
	} else {
		totalMemory := float64(vMemory.Total)
		freeMemory := float64(vMemory.Free)
		out = append(out,
			models.JSONMetric{ID: "TotalMemory", MType: gauge, Value: &totalMemory},
			models.JSONMetric{ID: "FreeMemory", MType: gauge, Value: &freeMemory},
		)
	}

	l.Info("geting CPU utilization info")
	utils, err := cpu.Percent(0, true)
	if err != nil {
		l.Error("getting CPU utilization", "error", err.Error())
		errs = errors.Join(errs, myErrors.ErrGetCPUutilization, err)
	} else {
		for i, util := range utils {
			metricName := fmt.Sprintf("CPUutilization%d", i)
			u := util
			out = append(out, models.JSONMetric{ID: metricName, MType: gauge, Value: &u})
		}
	}

	return out, errs
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

	rv := rand.Float64()
	out = append(out, models.JSONMetric{
		ID:    "RandomValue",
		MType: gauge,
		Value: &rv,
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
func PollMetrics(wg *sync.WaitGroup, controlChan chan bool, dataChan chan MetricsAddGetter, l logger, config *config.AgentConfig) {
	defer wg.Done()
	var pollCounter int64 = 0
POLLING:
	for {
		controlChan <- false

		l.Info("Poll counter", "Value", pollCounter)
		ms, err := collectMetrics(l, pollCounter)
		if err != nil {
			l.Error("collecting metrics", "error", err.Error())
		}
		if len(dataChan) == cap(dataChan) {
			l.Error("Error internal commnication", "error", myErrors.ErrChannelFull.Error())
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
func ReportMetrics(wg *sync.WaitGroup, controlChan chan bool, dataChan chan MetricsAddGetter, l logger, conf *config.AgentConfig, client *http.Client) {
	defer wg.Done()
	var reportCounter uint64 = 0
REPORTING:
	for {
		controlChan <- false

		time.Sleep(conf.ReportInterval)
		for len(dataChan) > 0 {
			l.Info("Report counter", "Value", reportCounter)
			switch {
			case conf.ReportMode == `json` && conf.Compress == `gzip` && !conf.Batch:
				go func(l logger, dataChan chan MetricsAddGetter, config *config.AgentConfig, client *http.Client) {
					err := sendMetricaToServerJSONgzip(l, <-dataChan, config.AddressServer, client, conf.Key)
					if err != nil {
						l.Error("sending gzipped json metrica", "error", err.Error())
					}
				}(l, dataChan, conf, client)
			case conf.ReportMode == `json` && conf.Compress == `none` && !conf.Batch:
				go func(l logger, dataChan chan MetricsAddGetter, config *config.AgentConfig, client *http.Client) {
					err := sendMetricaToServerJSON(l, <-dataChan, config.AddressServer, client, conf.Key)
					if err != nil {
						l.Error("sending nongzipped json metrica", "error", err.Error())
					}
				}(l, dataChan, conf, client)
			case conf.ReportMode == `raw` && !conf.Batch:
				go func(l logger, dataChan chan MetricsAddGetter, config *config.AgentConfig, client *http.Client) {
					err := sendMetricsToServerQueryStr(l, <-dataChan, config.AddressServer, client)
					if err != nil {
						l.Error("sending query string metrica", "error", err.Error())
					}
				}(l, dataChan, conf, client)

			case conf.Batch && conf.Compress == `none`:
				go func(l logger, dataChan chan MetricsAddGetter, config *config.AgentConfig, client *http.Client) {
					err := sendMetricaToServerBatch(l, <-dataChan, config.AddressServer, client, conf.Key)
					if err != nil {
						l.Error("sending nongzipped batch of metrics", "error", err.Error())
					}
				}(l, dataChan, conf, client)

			case conf.Batch && conf.Compress == `gzip`:
				go func(l logger, dataChan chan MetricsAddGetter, config *config.AgentConfig, client *http.Client) {
					err := sendMetricaToServerBatchgzip(l, <-dataChan, config.AddressServer, client, conf.Key)
					if err != nil {
						l.Error("sending gzipped batch of metrics", "error", err.Error())
					}
				}(l, dataChan, conf, client)
			}
		}
		reportCounter++

		if <-controlChan {
			l.Info("Reporting metrica stopped")
			break REPORTING
		}
	}
}
