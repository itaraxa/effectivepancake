package config

import (
	"flag"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/itaraxa/effectivepancake/internal/version"
)

type AgentConfig struct {
	PollInterval   time.Duration
	ReportInterval time.Duration
	AddressServer  string
	LogLevel       string
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		AddressServer:  `localhost:8080`,
		LogLevel:       `INFO`,
	}
}

func (ac *AgentConfig) ParseFlags() {
	var showVersion bool
	flag.BoolVar(&showVersion, `v`, false, `Show version and exit`)
	flag.StringVar(&ac.AddressServer, `a`, `localhost:8080`, `HTTP-server endpoint address`)
	var p, r int64
	flag.Int64Var(&p, `p`, 2, `metrics poll interval, seconds`)
	flag.Int64Var(&r, `r`, 10, `metrics report interval, seconds`)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\nUsage of %s\n", version.AgentVersion, os.Args[0])
		flag.PrintDefaults()
	}

	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v", err)
		os.Exit(1)
	}

	ac.PollInterval = time.Duration(p) * time.Second
	ac.ReportInterval = time.Duration(r) * time.Second

	if showVersion {
		fmt.Printf("Agent version: %s\n\r", version.AgentVersion)
		os.Exit(0)
	}
}

func (ac *AgentConfig) ParseEnv() {
	p, ok := os.LookupEnv(`POLL_INTERVAL`)
	if ok {
		pi, err := strconv.Atoi(p)
		if err != nil {
			ac.PollInterval = time.Duration(pi) * time.Second
		}
	}

	r, ok := os.LookupEnv(`REPORT_INTERVAL`)
	if ok {
		ri, err := strconv.Atoi(r)
		if err != nil {
			ac.ReportInterval = time.Duration(ri) * time.Second
		}
	}

	addressServer, ok := os.LookupEnv(`ADDRESS`)
	if ok {
		ac.AddressServer = addressServer
	}
}

type ServerConfig struct {
	endpoint string
}
