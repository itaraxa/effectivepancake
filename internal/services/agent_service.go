package services

import (
	"fmt"
	"math/rand"
	"runtime"

	"github.com/itaraxa/effectivepancake/internal/models"
)

func SendMetricsToServer(ms *models.Metrics, serverURL string) error {
	fmt.Printf("Server URL = %s\n\r", serverURL)
	fmt.Println(ms.String())

	// req, err := http.NewRequest("POST", fmt.Sprintf("http://localhost:8080/update/counter/%s/%d", "AgentTest", count), bytes.NewBuffer([]byte(``)))
	// if err != nil {
	// 	fmt.Printf("Error: %v", errors.ErrRequestCreating)
	// 	return
	// }

	// req.Header.Set("Content-Type", "txt/plain")

	// client := &http.Client{}
	// resp, err := client.Do(req)
	// if err != nil {
	// 	fmt.Printf("Error: %v", errors.ErrRequestSending)
	// 	return
	// }

	// if resp.StatusCode == http.StatusOK {
	// 	fmt.Println("Status code = 200. Succes request")
	// } else {
	// 	fmt.Printf("Status code = %d", resp.StatusCode)
	// }

	// defer resp.Body.Close()

	return nil
}

func CollectMetrics(pollCount uint64) (*models.Metrics, error) {
	ms := &models.Metrics{}

	ms.AddPollCount(pollCount)
	ms.AddData(collectRuntimeMetrics())
	ms.AddData(collectOtherMetrics())

	return ms, nil
}

func collectRuntimeMetrics() []models.Metrica {
	var memStats runtime.MemStats
	runtime.ReadMemStats(&memStats)

	out := []models.Metrica{}
	out = append(out, models.Metrica{Name: "Alloc", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Alloc)})
	out = append(out, models.Metrica{Name: "BuckHashSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.BuckHashSys)})
	out = append(out, models.Metrica{Name: "Frees", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Frees)})
	out = append(out, models.Metrica{Name: "GCCPUFraction", Type: "gauge", Value: fmt.Sprintf("%f", memStats.GCCPUFraction)})
	out = append(out, models.Metrica{Name: "GCSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.GCSys)})
	out = append(out, models.Metrica{Name: "HeapAlloc", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapAlloc)})
	out = append(out, models.Metrica{Name: "HeapIdle", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapIdle)})
	out = append(out, models.Metrica{Name: "HeapInuse", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapInuse)})
	out = append(out, models.Metrica{Name: "HeapObjects", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapObjects)})
	out = append(out, models.Metrica{Name: "HeapReleased", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapReleased)})
	out = append(out, models.Metrica{Name: "HeapSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.HeapSys)})
	out = append(out, models.Metrica{Name: "LastGC", Type: "gauge", Value: fmt.Sprintf("%d", memStats.LastGC)})
	out = append(out, models.Metrica{Name: "Lookups", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Lookups)})
	out = append(out, models.Metrica{Name: "MCacheInuse", Type: "gauge", Value: fmt.Sprintf("%d", memStats.MCacheInuse)})
	out = append(out, models.Metrica{Name: "MCacheSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.MCacheSys)})
	out = append(out, models.Metrica{Name: "MSpanInuse", Type: "gauge", Value: fmt.Sprintf("%d", memStats.MSpanInuse)})
	out = append(out, models.Metrica{Name: "MSpanSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.MSpanSys)})
	out = append(out, models.Metrica{Name: "Mallocs", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Mallocs)})
	out = append(out, models.Metrica{Name: "NextGC", Type: "gauge", Value: fmt.Sprintf("%d", memStats.NextGC)})
	out = append(out, models.Metrica{Name: "NumForcedGC", Type: "gauge", Value: fmt.Sprintf("%d", memStats.NumForcedGC)})
	out = append(out, models.Metrica{Name: "NumGC", Type: "gauge", Value: fmt.Sprintf("%d", memStats.NumGC)})
	out = append(out, models.Metrica{Name: "OtherSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.OtherSys)})
	out = append(out, models.Metrica{Name: "PauseTotalNs", Type: "gauge", Value: fmt.Sprintf("%d", memStats.PauseTotalNs)})
	out = append(out, models.Metrica{Name: "StackInuse", Type: "gauge", Value: fmt.Sprintf("%d", memStats.StackInuse)})
	out = append(out, models.Metrica{Name: "StackSys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.StackSys)})
	out = append(out, models.Metrica{Name: "Sys", Type: "gauge", Value: fmt.Sprintf("%d", memStats.Sys)})
	out = append(out, models.Metrica{Name: "TotalAlloc", Type: "gauge", Value: fmt.Sprintf("%d", memStats.TotalAlloc)})

	return out
}

func collectOtherMetrics() []models.Metrica {
	out := []models.Metrica{}

	out = append(out, models.Metrica{
		Name:  "RandomValue",
		Type:  "gauge",
		Value: fmt.Sprintf("%f", rand.Float64()),
	})

	return out
}
