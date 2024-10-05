package main

import (
	"flag"
	"fmt"
	"os"
	"time"
)

var config struct {
	pollInterval   time.Duration
	reportInterval time.Duration
	addressServer  string
}

func parseFlags() {
	var showVersion bool
	flag.BoolVar(&showVersion, `v`, false, `Show version and exit`)
	flag.StringVar(&config.addressServer, `a`, `localhost:8080`, `HTTP-server endpoint address`)
	flag.DurationVar(&config.pollInterval, `p`, 2*time.Second, `metrics poll interval`)
	flag.DurationVar(&config.reportInterval, `r`, 10*time.Second, `metrics report interval`)

	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\nUsage of %s\n", version, os.Args[0])
		flag.PrintDefaults()
	}

	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v", err)
		os.Exit(1)
	}
	if showVersion {
		fmt.Printf("Agent version: %s\n\r", version)
		os.Exit(0)
	}
}
