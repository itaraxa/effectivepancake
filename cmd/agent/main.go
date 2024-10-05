package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
	"github.com/itaraxa/effectivepancake/internal/services"
)

var version string = "1.5.0"

type AgentApp struct {
	logger     *slog.Logger
	httpClient *http.Client
	config     *struct {
		pollInterval   time.Duration
		reportInterval time.Duration
		addressServer  string
	}
	wg *sync.WaitGroup
}

func NewAgentApp(logger *slog.Logger, httpClient *http.Client, config *struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	addressServer  string
}) *AgentApp {
	return &AgentApp{
		logger:     logger,
		httpClient: httpClient,
		config:     config,
		wg:         new(sync.WaitGroup),
	}
}

func (aa *AgentApp) Run() {
	aa.logger.Info("Agent started", slog.String("server", aa.config.addressServer))
	defer aa.logger.Info("Agent stopped")

	var wg sync.WaitGroup
	msCh := make(chan *models.Metrics, aa.config.reportInterval/aa.config.pollInterval+1) // создаем канал для обмена данными между сборщиком и отправщиком
	defer close(msCh)

	// Ctrl+C handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	go func() {
		<-signalChan
		// Wait for closing requests
		time.Sleep(1 * time.Second)
		aa.logger.Info("Agent stopped", slog.String("reason", "Ctrl+C press"))
		os.Exit(0)
	}()

	// goroutine для сбора метрик
	wg.Add(1)
	go func(pollInterval time.Duration) {
		defer wg.Done()
		var pollCounter uint64 = 0
		for {
			aa.logger.Debug("Poll counter", slog.Uint64("Value", pollCounter))
			ms, err := services.CollectMetrics(pollCounter)
			if err != nil {
				aa.logger.Error("Error collect metrics", slog.String("error", err.Error()))
			}
			if len(msCh) == cap(msCh) {
				aa.logger.Error("Error internal commnication", slog.String("error", errors.ErrChannelFull.Error()))
			}
			msCh <- ms
			pollCounter += 1
			time.Sleep(pollInterval)
		}
	}(aa.config.pollInterval)

	// goroutine для отправки метрик
	wg.Add(1)
	go func(reportInterval time.Duration) {
		defer wg.Done()
		var reportCounter uint64 = 0
		for {
			time.Sleep(reportInterval)
			for len(msCh) > 0 {
				aa.logger.Debug("Report counter", slog.Uint64("Value", reportCounter))
				err := services.SendMetricsToServer(<-msCh, aa.config.addressServer, aa.httpClient)
				if err != nil {
					aa.logger.Error("Error sending to server. Waiting 1 second",
						slog.String("server", aa.config.addressServer),
						slog.String("error", errors.ErrSendingMetricsToServer.Error()),
					)
					// Pause for next sending
					time.Sleep(1 * time.Second)
				}
			}
			reportCounter++
		}
	}(aa.config.reportInterval)

	wg.Wait()
	aa.logger.Info("Agent stopped")
}

func main() {
	parseFlags()
	parseEnv()

	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	// config := struct {
	// 	pollInterval   time.Duration
	// 	reportInterval time.Duration
	// 	addressServer  string
	// }{
	// 	pollInterval:   1 * time.Second,
	// 	reportInterval: 2 * time.Second,
	// 	addressServer:  `localhost:8080`,
	// }
	myClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	app := NewAgentApp(logger, myClient, &config)
	app.Run()
}
