package main

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
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

// TO-DO: rewrite with interfaces
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
	)
	defer aa.logger.Info("Agent stopped")

	var wg sync.WaitGroup
	msCh := make(chan services.MetricSender, aa.config.ReportInterval/aa.config.PollInterval+1) // создаем канал для обмена данными между сборщиком и отправщиком
	defer close(msCh)

	// Ctrl+C handling
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)
	pollingStopChan := make(chan bool, 1)
	reportStopChan := make(chan bool, 1)
	go func() {
		<-signalChan
		aa.logger.Info("Agent stopping", "reason", "Ctrl+C press")
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

	logger, err := logger.NewZapLogger(agentConf.LogLevel)
	if err != nil {
		panic(err)
	}
	defer logger.Sync()
	myClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	app := NewAgentApp(logger, myClient, agentConf)
	app.Run()
}
