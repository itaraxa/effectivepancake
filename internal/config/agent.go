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
	ShowVersion    bool
	ReportMode     string // json or raw
	Compress       string // gzip or none
	Batch          bool
}

func NewAgentConfig() *AgentConfig {
	return &AgentConfig{
		PollInterval:   2 * time.Second,
		ReportInterval: 10 * time.Second,
		AddressServer:  `localhost:8080`,
		LogLevel:       `INFO`,
		ReportMode:     `json`,
		Compress:       `gzip`,
		Batch:          true,
	}
}

func (ac *AgentConfig) ParseFlags() error {
	flag.BoolVar(&ac.ShowVersion, `v`, false, `Show version and exit`)
	flag.BoolVar(&ac.Batch, `b`, true, `Use batch mode`)
	flag.StringVar(&ac.AddressServer, `a`, `localhost:8080`, `HTTP-server endpoint address. Environment variable ADDRESS`)
	flag.StringVar(&ac.LogLevel, `log`, `INFO`, `Set log level: INFO, DEBUG, etc. `)
	flag.StringVar(&ac.ReportMode, `m`, `json`, `Set method to report metrics: json, raw. Environment variable REPORT_METHOD`)
	flag.StringVar(&ac.Compress, `c`, `gzip`, `Set a data compression method: gzip or none. Environment variable COMPRESS`)
	var p, r int64
	flag.Int64Var(&p, `p`, 2, `metrics poll interval, seconds. Environment variable POLL_INTERVAL`)
	flag.Int64Var(&r, `r`, 10, `metrics report interval, seconds. Environment variable REPORT_INTERVAL`)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\nUsage of %s\n", version.AgentVersion, os.Args[0])
		flag.PrintDefaults()
	}

	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v", err)
		return err
	}

	ac.PollInterval = time.Duration(p) * time.Second
	ac.ReportInterval = time.Duration(r) * time.Second

	return nil
}

func (ac *AgentConfig) ParseEnv() error {
	p, ok := os.LookupEnv(`POLL_INTERVAL`)
	if ok {
		pi, err := strconv.Atoi(p)
		if err != nil {
			return err
		}
		ac.PollInterval = time.Duration(pi) * time.Second
	}

	r, ok := os.LookupEnv(`REPORT_INTERVAL`)
	if ok {
		ri, err := strconv.Atoi(r)
		if err != nil {
			return err
		}
		ac.ReportInterval = time.Duration(ri) * time.Second
	}

	addressServer, ok := os.LookupEnv(`ADDRESS`)
	if ok {
		ac.AddressServer = addressServer
	}

	l, ok := os.LookupEnv(`LOG_LEVEL`)
	if ok {
		switch l {
		case `INFO`:
			ac.LogLevel = `INFO`
		case `DEBUG`:
			ac.LogLevel = `DEBUG`
		case `ERROR`:
			ac.LogLevel = `ERROR`
		default:
			fmt.Fprintln(os.Stderr, `An unknown log level was set via an environment variable. Using INFO`)
			ac.LogLevel = `INFO`
		}
	}

	m, ok := os.LookupEnv(`REPORT_MODE`)
	if ok {
		ac.ReportMode = m
	}

	c, ok := os.LookupEnv(`COMPRESS`)
	if ok {
		switch c {
		case `gzip`:
			ac.Compress = `gzip`
		case `none`:
			ac.Compress = `none`
		default:
			fmt.Fprintln(os.Stderr, `An unknown compress method was set via an environment variable. Using gzip`)
			ac.Compress = `gzip`
		}
	}
	return nil
}
