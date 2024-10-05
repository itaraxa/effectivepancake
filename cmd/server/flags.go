package main

import (
	"flag"
	"fmt"
	"os"
)

var config struct {
	endpoint string
}

func parseFlags() {
	var showVersion bool
	flag.BoolVar(&showVersion, `v`, false, `Show version and exit`)
	flag.StringVar(&config.endpoint, `a`, `localhost:8080`, `HTTP-server endpoint address`)
	flag.Usage = func() {
		fmt.Fprintf(flag.CommandLine.Output(), "Version: %s\nUsage of %s\n", version, os.Args[0])
		flag.PrintDefaults()
	}
	// Вместо flag.Parse() использую flag.CommandLine.Parse(os.Args[1:]) для возврата ошибки и проверки флагов
	err := flag.CommandLine.Parse(os.Args[1:])
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error parsing flags: %v", err)
		os.Exit(1)
	}
	if showVersion {
		fmt.Printf("Server version: %s\n\r", version)
		os.Exit(0)
	}
}
