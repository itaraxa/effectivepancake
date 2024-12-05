package services

import (
	"fmt"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/itaraxa/effectivepancake/internal/config"
)

func ReportMetricsDispatcher(gwg *sync.WaitGroup, jobCount, workerCount int, l logger, config *config.AgentConfig, httpClient *http.Client, msCh chan MetricsAddGetter) {
	defer gwg.Done()
	l.Info("starting metrics report dispatcher")
	jobs := make(chan Job, jobCount)
	results := make(chan Result, jobCount)

	// control channel capacity
	control := make(chan struct{})
	defer close(control)
	go channelStats(l, control, jobs, results, msCh)

	var wg sync.WaitGroup

	// start workers
	wg.Add(workerCount)
	for w := 1; w <= workerCount; w++ {
		l.Info("creating worker", "workerID", w)
		go worker(l, w, jobs, results, &wg, config, httpClient)
	}

	// start result collector
	var resultWg sync.WaitGroup
	resultWg.Add(1)
	go collectResults(results, &resultWg, l)

	var jobCounter atomic.Uint64
	jobCounter.Store(0)
	// start job distribution
	for metric := range msCh {
		jobs <- Job{JobID: jobCounter.Load(), Metrics: []MetricsAddGetter{metric}}
		jobCounter.Add(1)
	}
	l.Info("channel msCh closed")
	close(jobs)
	l.Info("channel jobs closed")
	wg.Wait()
	close(results)
	l.Info("channel results closed")
	resultWg.Wait()
}

type Job struct {
	JobID   uint64
	Metrics []MetricsAddGetter
}

type Result struct {
	JobID    uint64
	WorkerID int
	Result   error
}

func worker(l logger, id int, jobs <-chan Job, result chan<- Result, wg *sync.WaitGroup, config *config.AgentConfig, httpClient *http.Client) {
	defer wg.Done()

	l.Info("start worker", "workerID", id)

	for job := range jobs {
		l.Debug("adding new job into worker", "jobID", job.JobID, "workerID", id)
		result <- Result{JobID: job.JobID, WorkerID: id, Result: reportMetricsWorker(job.Metrics, l, config, httpClient)}
	}
	l.Info("stop worker", "workerID", id)
}

func collectResults(results <-chan Result, wg *sync.WaitGroup, l logger) {
	defer wg.Done()
	for result := range results {
		l.Debug("get result", "jobID", result.JobID, "workerID", result.WorkerID)
		if result.Result != nil {
			l.Error("worker sending data", "workerID", result.WorkerID, "result", "failure", "error", fmt.Sprintf("Job ID: %d, error: %s", result.JobID, result.Result))
		} else {
			l.Info("worker sending data", "workerID", result.WorkerID, "result", "done")
		}
	}
	l.Info("stop collecting results")
}

func reportMetricsWorker(metrics []MetricsAddGetter, l logger, conf *config.AgentConfig, client *http.Client) error {
	var reportCounter uint64 = 0

	for _, metric := range metrics {
		l.Info("sleeping", "time", conf.ReportInterval)
		time.Sleep(conf.ReportInterval)
		switch {
		case conf.ReportMode == `json` && conf.Compress == `gzip` && !conf.Batch:
			err := sendMetricaToServerJSONgzip(l, metric, conf.AddressServer, client, conf.Key)
			if err != nil {
				l.Error("sending gzipped json metrica", "error", err.Error())
				return err
			}

		case conf.ReportMode == `json` && conf.Compress == `none` && !conf.Batch:
			err := sendMetricaToServerJSON(l, metric, conf.AddressServer, client, conf.Key)
			if err != nil {
				l.Error("sending nongzipped json metrica", "error", err.Error())
				return err
			}

		case conf.ReportMode == `raw` && !conf.Batch:
			err := sendMetricsToServerQueryStr(l, metric, conf.AddressServer, client)
			if err != nil {
				l.Error("sending query string metrica", "error", err.Error())
				return err
			}

		case conf.Batch && conf.Compress == `none`:
			err := sendMetricaToServerBatch(l, metric, conf.AddressServer, client, conf.Key)
			if err != nil {
				l.Error("sending nongzipped batch of metrics", "error", err.Error())
				return err
			}

		case conf.Batch && conf.Compress == `gzip`:
			err := sendMetricaToServerBatchgzip(l, metric, conf.AddressServer, client, conf.Key)
			if err != nil {
				l.Error("sending gzipped batch of metrics", "error", err.Error())
				return err
			}
		}
		reportCounter++
	}
	return nil
}

func channelStats(l logger, stopCh chan struct{}, jobs chan Job, results chan Result, msCh chan MetricsAddGetter) {
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	l.Debug("start collecting channel stats")
	for {
		select {
		case <-stopCh:
			l.Debug("finish collecting channel stats")
			return
		case <-ticker.C:
			l.Debug("jobs channel", "len", len(jobs), "cap", cap(jobs))
			l.Debug("results channel", "len", len(results), "cap", cap(results))
			l.Debug("msCh channel", "len", len(msCh), "cap", cap(msCh))
		}
	}
}
