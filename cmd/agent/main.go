package main

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/itaraxa/effectivepancake/internal/errors"
	"github.com/itaraxa/effectivepancake/internal/models"
	"github.com/itaraxa/effectivepancake/internal/services"
)

func main() {
	pollInterval := 1 * time.Second
	reportInterval := 4 * time.Second
	var wg sync.WaitGroup
	msCh := make(chan *models.Metrics, reportInterval/pollInterval+1) // создаем канала для обмена данными между сборщиком и отправщиком

	myClient := &http.Client{
		Timeout: 1 * time.Second,
	}

	// goroutine для сбора метрик
	wg.Add(1)
	go func(pollInterval time.Duration) {
		defer wg.Done()
		var pollCounter uint64 = 0
		for {
			fmt.Printf("pollCounter   = %d\n\r", pollCounter)
			ms, err := services.CollectMetrics(pollCounter)
			if err != nil {
				return
			}
			if len(msCh) == cap(msCh) {
				fmt.Printf("error: %v\n\r", errors.ErrChannelFull)
			}
			msCh <- ms
			pollCounter += 1
			time.Sleep(pollInterval)
		}
	}(pollInterval)

	// goroutine для отправки метрик
	wg.Add(1)
	go func(reportInterval time.Duration) {
		defer wg.Done()
		var reportCounter uint64 = 0
		for {
			time.Sleep(reportInterval)
			for len(msCh) > 0 {
				fmt.Printf("reportCounter = %d\n\r", reportCounter)
				err := services.SendMetricsToServer(<-msCh, "localhost:8080", myClient)
				if err != nil {
					fmt.Printf("Sending error: %v", errors.ErrSendingMetricsToServer)
					return
				}
				// fmt.Println(<-msCh)
			}
			reportCounter++
		}
	}(reportInterval)

	wg.Wait()
	fmt.Printf("EXIT")

}
