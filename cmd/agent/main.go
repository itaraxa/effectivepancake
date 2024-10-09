package main

import (
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/itaraxa/effectivepancake/internal/config"
	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
	"github.com/itaraxa/effectivepancake/internal/services"
	"github.com/itaraxa/effectivepancake/internal/version"
)

type AgentApp struct {
	logger     *slog.Logger
	httpClient *http.Client
	config     *config.AgentConfig
	wg         *sync.WaitGroup
}

// TO-DO: rewrite with interfaces
func NewAgentApp(logger *slog.Logger, httpClient *http.Client, config *config.AgentConfig) *AgentApp {
	return &AgentApp{
		logger:     logger,
		httpClient: httpClient,
		config:     config,
		wg:         new(sync.WaitGroup),
	}
}

func (aa *AgentApp) Run() {
	aa.logger.Info("Agent version", slog.String("Version", version.AgentVersion))
	aa.logger.Info("Agent started",
		slog.String("server", aa.config.AddressServer),
		slog.Duration("poll interval", aa.config.PollInterval),
		slog.Duration("report interval", aa.config.ReportInterval),
	)
	defer aa.logger.Info("Agent stopped")

	var wg sync.WaitGroup
	msCh := make(chan *models.Metrics, aa.config.ReportInterval/aa.config.PollInterval+1) // создаем канал для обмена данными между сборщиком и отправщиком
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
	}(aa.config.PollInterval)

	// goroutine для отправки метрик
	wg.Add(1)
	go func(reportInterval time.Duration) {
		defer wg.Done()
		var reportCounter uint64 = 0
		for {
			time.Sleep(reportInterval)
			for len(msCh) > 0 {
				aa.logger.Debug("Report counter", slog.Uint64("Value", reportCounter))
				err := services.SendMetricsToServer(<-msCh, aa.config.AddressServer, aa.httpClient)
				if err != nil {
					aa.logger.Error("Error sending to server. Waiting 1 second",
						slog.String("server", aa.config.AddressServer),
						slog.String("error", errors.ErrSendingMetricsToServer.Error()),
					)
					// Pause for next sending
					time.Sleep(1 * time.Second)
				}
			}
			reportCounter++
		}
	}(aa.config.ReportInterval)

	wg.Wait()
	aa.logger.Info("Agent stopped")
}

func main() {
	// Preparing the configuration for Agent app startup
	agentConf := config.NewAgentConfig()
	agentConf.ParseFlags()
	agentConf.ParseEnv()

	// Creating dependencies
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	myClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	app := NewAgentApp(logger, myClient, agentConf)
	app.Run()
}
