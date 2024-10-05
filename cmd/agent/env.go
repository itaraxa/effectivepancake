package main

import (
	"os"
	"strconv"
	"time"
)

func parseEnv() {
	p, ok := os.LookupEnv(`POLL_INTERVAL`)
	if ok {
		pi, err := strconv.Atoi(p)
		if err != nil {
			config.pollInterval = time.Duration(pi) * time.Second
		}
	}

	r, ok := os.LookupEnv(`REPORT_INTERVAL`)
	if ok {
		ri, err := strconv.Atoi(r)
		if err != nil {
			config.pollInterval = time.Duration(ri) * time.Second
		}
	}

	addressServer, ok := os.LookupEnv(`ADDRESS`)
	if ok {
		config.addressServer = addressServer
	}
}
