package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/itaraxa/effectivepancake/internal/repositories/agentbuffer"
	"github.com/itaraxa/effectivepancake/internal/services"
)

func main() {
	pollInterval := 2 * time.Second
	reportInterval := 10 * time.Second
	var wg sync.WaitGroup

	myClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	// буфер для накопления метрик на агенте
	ab := new(agentbuffer.AgentBuffer)

	// goroutine для сбора метрик
	wg.Add(1)
	go func(pollInterval time.Duration) {
		defer wg.Done()
		var pollCounter uint64 = 0
		for {
			_, err := services.CollectMetrics(pollCounter)
			if err != nil {
				return
			}
			time.Sleep(pollInterval)
		}
	}(pollInterval)

	// goroutine для отправки метрик
	wg.Add(1)
	go func(reportInterval time.Duration, client *http.Client, ab *agentbuffer.AgentBuffer) {
		defer wg.Done()

		err := services.SendMetricsToServer(nil, "localhost:8080")
		if err != nil {
			return
		}
		time.Sleep(reportInterval)
	}(reportInterval, myClient, ab)

	time.Sleep(reportInterval)
	wg.Wait()
	fmt.Printf("EXIT")

}
