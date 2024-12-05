package errors

import "errors"

var (
	// Server errors
	ErrParseGauge              = errors.New("gauge value parsing error")
	ErrParseCounter            = errors.New("counter value parsing error")
	ErrBadType                 = errors.New("bad type in query for update")
	ErrBadName                 = errors.New("bad name in raw query string")
	ErrBadValue                = errors.New("bad value in raw query string")
	ErrBadRawQuery             = errors.New("bad raw query string")
	ErrMetricaNotFaund         = errors.New("unknown metrica")
	ErrUpdateGauge             = errors.New("gauge value updating error")
	ErrAddCounter              = errors.New("counter value adding error")
	ErrEmptyMetricaName        = errors.New("empty metric name in query")
	ErrEmptyMetricaRawValue    = errors.New("empty metrica raw value in query")
	ErrGettingAnswerFromServer = errors.New("cannot read server answer")
	ErrMemStorageNotInitilized = errors.New("memstorage not initialized")
	ErrGetVirtualMemory        = errors.New("cannot get virtual memory")
	ErrGetCPUutilization       = errors.New("cannot get CPU utilization")

	// Agent errors
	ErrRequestCreating        = errors.New("creating request error")
	ErrRequestSending         = errors.New("sending request error")
	ErrBufferIsEmpty          = errors.New("buffer on agent is empty")
	ErrChannelFull            = errors.New("channel in the agent is full")
	ErrSendingMetricsToServer = errors.New("error in services.SendMetricsToServer")
	ErrAddPollCount           = errors.New("error adding pollCount into an instance of the models.Metrics")
	ErrAddData                = errors.New("error adding data into an instance of the models.Metrics")
	ErrNoMetrics              = errors.New("GetData() didn't return any metrics")
)
