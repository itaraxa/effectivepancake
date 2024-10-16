package main

import (
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/itaraxa/effectivepancake/internal/config"
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
	msCh := make(chan services.Metricer, aa.config.ReportInterval/aa.config.PollInterval+1) // создаем канал для обмена данными между сборщиком и отправщиком
	defer close(msCh)

	// Ctrl+C handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	pollingStopChan := make(chan bool, 1)
	reportStopChan := make(chan bool, 1)
	go func() {
		<-signalChan
		aa.logger.Info("Agent stopping", slog.String("reason", "Ctrl+C press"))
		// sending true to the control channel if polling/reporting should stop
		// reading from channel for unblocking
		<-pollingStopChan
		pollingStopChan <- true
		<-reportStopChan
		reportStopChan <- true
	}()

	// goroutine для сбора метрик
	wg.Add(1)
	go services.PollMetrics(&wg, pollingStopChan, msCh, aa.logger, aa.config)

	// goroutine для отправки метрик
	wg.Add(1)
	go services.ReportMetrics(&wg, reportStopChan, msCh, aa.logger, aa.config, aa.httpClient)

	wg.Wait()
}

func main() {
	// Preparing the configuration for Agent app startup
	agentConf := config.NewAgentConfig()
	err := agentConf.ParseFlags()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	if agentConf.ShowVersion {
		fmt.Println(version.AgentVersion)
		os.Exit(0)
	}

	err = agentConf.ParseEnv()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	var level slog.Level
	switch agentConf.LogLevel {
	case "DEBUG":
		level = slog.LevelDebug
	case "INFO":
		level = slog.LevelInfo
	case "WARN":
		level = slog.LevelWarn
	case "ERROR":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}
	logger := slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level}))

	myClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	app := NewAgentApp(logger, myClient, agentConf)
	app.Run()
}
