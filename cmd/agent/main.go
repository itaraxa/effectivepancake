package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"time"

	"github.com/itaraxa/effectivepancake/internal/config"
	"github.com/itaraxa/effectivepancake/internal/logger"
	"github.com/itaraxa/effectivepancake/internal/services"
	"github.com/itaraxa/effectivepancake/internal/version"
)

type AgentApp struct {
	logger     logger.Logger
	httpClient *http.Client
	config     *config.AgentConfig
	wg         *sync.WaitGroup
}

func NewAgentApp(logger logger.Logger, httpClient *http.Client, config *config.AgentConfig) *AgentApp {
	return &AgentApp{
		logger:     logger,
		httpClient: httpClient,
		config:     config,
		wg:         new(sync.WaitGroup),
	}
}

func (aa *AgentApp) Run() {
	aa.logger.Info("Agent version", "Version", version.AgentVersion)
	aa.logger.Info("Agent started", "server", aa.config.AddressServer,
		"poll interval", aa.config.PollInterval,
		"report interval", aa.config.ReportInterval,
		"log level", aa.config.LogLevel,
		"report mode", aa.config.ReportMode,
		"compress methode", aa.config.Compress,
		"batch mode", aa.config.Batch,
		"key", aa.config.Key,
		"report rate limit", aa.config.RateLimit,
	)
	defer aa.logger.Info("Agent stopped")

	var wg sync.WaitGroup
	msCh := make(chan services.MetricsAddGetter, 128) // создаем канал для обмена данными между сборщиком и отправщиком

	// Ctrl+C handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	stopChan := make(chan struct{})
	go func() {
		<-signalChan
		aa.logger.Info("Agent stopping", "reason", "Ctrl+C press")
		close(stopChan)
		close(msCh)
	}()

	// goroutine для сбора метрик
	wg.Add(1)
	go services.PollMetrics(&wg, stopChan, msCh, aa.logger, aa.config)

	// worker pool для отправки метрик
	wg.Add(1)
	if aa.config.RateLimit != 0 {
		// if RateLimit > 0 -> start worker pool
		aa.logger.Info("start worker pool for metrics reporting")
		go services.ReportMetricsDispatcher(&wg, cap(msCh), aa.config.RateLimit, aa.logger, aa.config, aa.httpClient, msCh)
	} else {
		// if RateLimit == 0 -> start single goroutine
		aa.logger.Info("start single reporting goroutine")
		go services.ReportMetrics(&wg, stopChan, msCh, aa.logger, aa.config, aa.httpClient)
	}

	wg.Add(1)
	go func(wg *sync.WaitGroup, stopCh chan struct{}) {
		defer wg.Done()
		ticker := time.NewTicker(500 * time.Millisecond)
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				aa.logger.Debug("goroutines", "count", runtime.NumGoroutine())
			case <-stopCh:
				return
			}
		}
	}(&wg, stopChan)

	wg.Wait()
}

func main() {
	// Preparing the configuration for Agent app startup
	agentConf := config.NewAgentConfig()
	err := agentConf.ParseFlags()
	if err != nil {
		log.Fatalf("error parsing comandline flags: %v", err.Error())
	}
	if agentConf.ShowVersion {
		fmt.Println(version.AgentVersion)
		return
	}

	err = agentConf.ParseEnv()
	if err != nil {
		log.Fatalf("error parsing environment variables: %v", err.Error())
	}

	logger, err := logger.NewZapLogger(agentConf.LogLevel)
	if err != nil {
		log.Fatalf("дogger initialization error: %v", err.Error())
	}
	defer logger.Sync()
	myClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	app := NewAgentApp(logger, myClient, agentConf)
	app.Run()
}
