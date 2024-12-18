package services

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	myErrors "github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

const (
	gauge   = `gauge`
	counter = `counter`
)

/*
Creating new instance of models.Query from request.URL string

Args:

	raw string: request.URL in string format. Example: "/update/gauge/test1/3.14"

Returns:

	q Querier: copy of instance, implemented Querier
	err error: nil or error occurred while parsing the raw string
*/
func ParseQueryString(raw string) (q Querier, err error) {
	queryString := raw[1:]
	if len(strings.Split(queryString, `/`)) != 4 {
		return nil, myErrors.ErrBadRawQuery
	}
	q = models.NewQuery()
	err = q.SetMetricaType(queryString)
	if err != nil {
		return models.NewQuery(), err
	}
	err = q.SetMetricaName(queryString)
	if err != nil {
		return models.NewQuery(), err
	}
	err = q.SetMetricaRawValue(queryString)
	if err != nil {
		return models.NewQuery(), err
	}
	return q, nil
}

/*
Writing data from the instance models.Query to storage.

Args:

	ctx context.Context
	q Querier: object, implementing Querier interface
	s Storager: object, implementing Storager interface

Returns:

	error: nil or error, if occurred
*/
func UpdateMetrica(ctx context.Context, q Querier, s MetricUpdater) error {
	switch q.GetMetricaType() {
	case gauge:
		g, err := strconv.ParseFloat(q.GetMetricaRawValue(), 64)
		if err != nil {
			return myErrors.ErrParseGauge
		}

		err = retryQueryToDB(func() error { return s.UpdateGauge(ctx, q.GetMetricName(), g) })

		if err != nil {
			return myErrors.ErrUpdateGauge
		}
	case counter:
		c, err := strconv.Atoi(q.GetMetricaRawValue())
		if err != nil {
			return myErrors.ErrParseCounter
		}
		err = retryQueryToDB(func() error { return s.AddCounter(ctx, q.GetMetricName(), int64(c)) })
		if err != nil {
			return myErrors.ErrAddCounter
		}
	default:
		return myErrors.ErrBadType
	}
	return nil
}

/*
JSONUpdateMetrica updates the metric value received from the request in the storage

Args:

	ctx context.Context
	jmq JSONMetricaQuerier: a request that allows getting the gauge value or counter delta
	mu MetricUpdater: a storage that allows saving a metric

Returns:

	error: nil or error, if occured
*/
func JSONUpdateMetrica(ctx context.Context, jmq JSONMetricaQuerier, mu MetricUpdater) error {
	ctx1s, cancel1s := context.WithTimeout(ctx, 1*time.Second)
	defer cancel1s()
	switch jmq.GetMetricaType() {
	case gauge:
		err := retryQueryToDB(func() error { return mu.UpdateGauge(ctx1s, jmq.GetMetricaName(), *jmq.GetMetricaValue()) })
		if err != nil {
			return myErrors.ErrUpdateGauge
		}
	case counter:
		err := retryQueryToDB(func() error { return mu.AddCounter(ctx1s, jmq.GetMetricaName(), *jmq.GetMetricaCounter()) })
		if err != nil {
			return myErrors.ErrAddCounter
		}
	default:
		return myErrors.ErrBadType
	}
	return nil
}

/*
WriteMetrics saves metrics from storage to the any writer

Args:

	ctx context.Context
	mg MetricGetter: a storage that allows getting metrics
	dst io.Writer: an object that allows data to be written to it

Returns:

	error: nil or error, if occured
*/
func WriteMetrics(ctx context.Context, mg MetricGetter, dst io.Writer) error {
	metrics, err := mg.GetAllMetrics(ctx)
	if err != nil {
		return err
	}
	data, err := json.MarshalIndent(metrics, "\t", "\t")
	if err != nil {
		return err
	}
	_, err = dst.Write(data)
	if err != nil {
		return err
	}
	return nil
}

/*
WriteMetrics saves metrics from storage and current timestamp to the any writer

Args:

	ctx context.Context
	mg MetricGetter: a storage that allows getting metrics
	dst io.Writer: an object that allows data to be written to it

Returns:

	error: nil or error, if occured
*/
func WriteMetricsWithTimestamp(ctx context.Context, mg MetricGetter, dst io.Writer) error {
	var err error
	blob := make(map[string]interface{})
	blob["timestamp"] = time.Now()
	blob["metrics"], err = mg.GetAllMetrics(ctx)
	if err != nil {
		return err
	}

	data, err := json.MarshalIndent(blob, "\t", "\t")
	if err != nil {
		return err
	}
	_, err = dst.Write(data)
	if err != nil {
		return err
	}
	return nil
}

/*
SaveMetricsToFile saves metrics from storage to the file.
THis is warapper over WriteMetricsWithTimestamp(mg MetricGetter, dst io.Writer) function

Args:

	ctx context.Context
	l logger: a logger used for printing messages
	mg MetricGetter: a storage that allows getting metrics
	fileName string: the file name for saving metric data

Returns:

	error: nil or error, if occured
*/
func SaveMetricsToFile(ctx context.Context, l logger, mg MetricGetter, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		l.Debug("cannot open file for writing", "error", err.Error(), "filename", fileName)
		return fmt.Errorf("cannot open %s for writing: %v", fileName, err)
	}
	defer file.Close()
	err = WriteMetricsWithTimestamp(ctx, mg, file)
	if err != nil {
		l.Debug("cannot save data to file", "error", err.Error(), "filename")
		return fmt.Errorf("cannot save date into %s: %v", fileName, err)
	}
	l.Info("file saved", "filename", fileName)
	return nil
}

/*
LoadMetrics loads metric data from any reader

Args:

	mu MetricUpdater: a storage that allows updating metric data
	src io.Reader: an object that allow reading data

Returns:

	time.Time: timestamp of when the metric data was written to the reader
	error: nil or error, if occured
*/
func LoadMetrics(mu MetricUpdater, src io.Reader) (time.Time, error) {
	data := make(map[string]interface{})
	decoder := json.NewDecoder(src)
	if err := decoder.Decode(&data); err != nil {
		return time.UnixMilli(0), fmt.Errorf("cannot unmarshal data")
	}
	timeStampStr, ok := data["timestamp"].(string)
	if !ok {
		return time.UnixMilli(0), fmt.Errorf("data doesn't contain timestamp field")
	}
	timeStamp, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", timeStampStr)
	if err != nil {
		return time.UnixMilli(0), fmt.Errorf("cann't parse timestamp: %v", err.Error())
	}
	metrics, ok := data["metrics"]
	if !ok {
		return time.UnixMilli(0), fmt.Errorf("data doesn't contain metrics")
	}

	if gauges, ok := metrics.(map[string]interface{})["gauges"]; ok {
		for ID, value := range gauges.(map[string]interface{}) {
			err = retryQueryToDB(func() error { return mu.UpdateGauge(context.TODO(), ID, value.(float64)) })
			if err != nil {
				return time.UnixMilli(0), fmt.Errorf("updating gauge %s error: %v", ID, err.Error())
			}
		}
	}
	if counter, ok := metrics.(map[string]interface{})["counters"]; ok {
		for ID, delta := range counter.(map[string]interface{}) {
			// Unmarshall from interface{} to float64 and convert to int64
			// because json.Unmarshall numbers into float64
			err = retryQueryToDB(func() error { return mu.AddCounter(context.TODO(), ID, int64(delta.(float64))) })
			if err != nil {
				return time.UnixMilli(0), fmt.Errorf("updating counter %s error: %v", ID, err.Error())
			}
		}
	}

	return timeStamp, nil
}

/*
LoadMetricsFromFile loades metric data from the file.
This is wrapper over LoadMetrics(mu MetricUpdater, src io.Reader) (time.Time, error) function

Args:

	l logger: a logger used for printing messages
	mu MetricUpdater: a storage that allows updating metric data
	fileName string: the file name for reading metric data

Returns:

	error: nil or error, if occured
*/
func LoadMetricsFromFile(l logger, mu MetricUpdater, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_RDONLY, 0666)
	if err != nil {
		l.Error("cannot open file for loading metrics", "error", err.Error(), "filename", fileName)
		return err
	}
	defer file.Close()
	l.Info("start loading metrics from file", "file name", fileName)
	timeStamp, err := LoadMetrics(mu, file)
	if err != nil {
		l.Error("cannot load metrics from file", "error", err.Error(), "filename", fileName)
		return err
	}
	l.Info("metrics have been loaded", "origin timestamp", timeStamp)

	return nil
}

/*
JSONUpdateBatchMetrica a function that performs batch updates of metrics in the storage

Args:

	ctx context.Context
	l logger: a logger used for printing messages
	jmqs []JSONMetricaQuerier: a slice of objcets, that implements the JSONMetricaQuerier interface
	mbu MetricBatchUpdater: object, that implements the MetricBatchUpdater interface

Returns:

	error
*/
func JSONUpdateBatchMetrica(ctx context.Context, l logger, jmqs []JSONMetricaQuerier, mbu MetricBatchUpdater) error {
	gaugeBatch := []struct {
		MetricName  string
		MetricValue *float64
	}{}

	counterBatch := []struct {
		MetricName  string
		MetricDelta *int64
	}{}

	for _, jmq := range jmqs {
		switch jmq.GetMetricaType() {
		case gauge:
			name := jmq.GetMetricaName()
			value := jmq.GetMetricaValue()
			gaugeBatch = append(gaugeBatch, struct {
				MetricName  string
				MetricValue *float64
			}{MetricName: name, MetricValue: value})

		case counter:
			name := jmq.GetMetricaName()
			delta := jmq.GetMetricaCounter()
			counterBatch = append(counterBatch, struct {
				MetricName  string
				MetricDelta *int64
			}{MetricName: name, MetricDelta: delta})
		}
	}
	l.Debug("get batch for load", "gauges", len(gaugeBatch), "counters", len(counterBatch))

	ctxWithTimeout, cancelWithTimeout := context.WithTimeout(ctx, 3*time.Second)
	defer cancelWithTimeout()

	err := retryQueryToDB(func() error { return mbu.UpdateBatchGauge(ctxWithTimeout, gaugeBatch) })
	if err != nil {
		l.Error("updating gauge batch", "error", err.Error())
		return err
	}

	err = retryQueryToDB(func() error { return mbu.AddBatchCounter(ctxWithTimeout, counterBatch) })
	if err != nil {
		l.Error("updating counter batch", "error", err.Error())
		return err
	}

	return nil
}
