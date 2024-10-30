package services

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
)

// Интерфейсы для описания взаимодействия с хранилищем метрик
type MetricStorager interface {
	MetricGetter
	MetricUpdater
	MetricPrinter
}

type MetricUpdater interface {
	UpdateGauge(metricName string, value float64) error
	AddCounter(metricName string, value int64) error
}

type MetricGetter interface {
	GetMetrica(metricaType string, metricaName string) (interface{}, error)
	GetAllMetrics() interface{}
}

type MetricPrinter interface {
	String() string
	HTML() string
}

// Интерфейс для описания взаимодействия с запросом на обновление метрики
// TO-DO: move from Qury to Metrica
type Querier interface {
	GetMetricaType() string
	SetMetricaType(string) error
	GetMetricName() string
	SetMetricaName(string) error
	GetMetricaRawValue() string
	SetMetricaRawValue(string) error
	String() string
}

type StringMetricaQuerier interface {
	SetMetricaRawValue(string) error
	String() string
}

type JSONMetricaQuerier interface {
	GetMetricaType() string
	GetMetricaName() string
	GetMetricaValue() *float64
	GetMetricaCounter() *int64
}

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
		return nil, errors.ErrBadRawQuery
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

	q Querier: object, implementing Querier interface
	s Storager: object, implementing Storager interface

Returns:

	error: nil or error, if occurred
*/
func UpdateMetrica(q Querier, s MetricUpdater) error {
	switch q.GetMetricaType() {
	case "gauge":
		g, err := strconv.ParseFloat(q.GetMetricaRawValue(), 64)
		if err != nil {
			return errors.ErrParseGauge
		}
		err = s.UpdateGauge(q.GetMetricName(), g)
		if err != nil {
			return errors.ErrUpdateGauge
		}
	case "counter":
		c, err := strconv.Atoi(q.GetMetricaRawValue())
		if err != nil {
			return errors.ErrParseCounter
		}
		err = s.AddCounter(q.GetMetricName(), int64(c))
		if err != nil {
			return errors.ErrAddCounter
		}
	default:
		return errors.ErrBadType
	}
	return nil
}

func JSONUpdateMetrica(jmq JSONMetricaQuerier, mu MetricUpdater) error {
	switch jmq.GetMetricaType() {
	case "gauge":
		err := mu.UpdateGauge(jmq.GetMetricaName(), *jmq.GetMetricaValue())
		if err != nil {
			return errors.ErrUpdateGauge
		}
	case "counter":
		err := mu.AddCounter(jmq.GetMetricaName(), *jmq.GetMetricaCounter())
		if err != nil {
			return errors.ErrAddCounter
		}
	default:
		return errors.ErrBadType
	}
	return nil
}

func SaveMetrics(mg MetricGetter, dst io.Writer) error {
	data, err := json.MarshalIndent(mg.GetAllMetrics(), "\t", "\t")
	if err != nil {
		return err
	}
	_, err = dst.Write(data)
	if err != nil {
		return err
	}
	return nil
}

func WriteMetricsWithTimestamp(mg MetricGetter, dst io.Writer) error {
	blob := make(map[string]interface{})
	blob["timestamp"] = time.Now()
	blob["metrics"] = mg.GetAllMetrics()

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

func SaveMetricsToFile(l logger, mg MetricGetter, fileName string) error {
	file, err := os.OpenFile(fileName, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, 0666)
	if err != nil {
		l.Debug("cannot open file for writing", "error", err.Error(), "filename", fileName)
		return fmt.Errorf("cannot open %s for writing: %v", fileName, err)
	}
	defer file.Close()
	err = WriteMetricsWithTimestamp(mg, file)
	if err != nil {
		l.Debug("cannot save data to file", "error", err.Error(), "filename")
		return fmt.Errorf("cannot save date into %s: %v", fileName, err)
	}
	l.Info("file saved", "filename", fileName)
	return nil
}

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
	timeStamp, err := time.Parse("2006-01-02T15:04:05.999999999-07:00", timeStampStr)
	if err != nil {
		return time.UnixMilli(0), fmt.Errorf("cann't parse timestamp: %v", err.Error())
	}
	metrics, ok := data["metrics"]
	if !ok {
		return time.UnixMilli(0), fmt.Errorf("data doesn't contain metrics")
	}

	if gauges, ok := metrics.(map[string]interface{})["gauges"]; ok {
		for ID, value := range gauges.(map[string]interface{}) {
			mu.UpdateGauge(ID, value.(float64))
		}
	}
	if counter, ok := metrics.(map[string]interface{})["counters"]; ok {
		for ID, delta := range counter.(map[string]interface{}) {
			// Unmarshall from interface{} to float64 and convert to int64
			// because json.Unmarshall numbers into float64
			mu.AddCounter(ID, int64(delta.(float64)))
		}
	}

	return timeStamp, nil
}

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
