package metrics

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
)

type MetricsStorage struct {
	RandomValue float64
	PollCount   int64
	Supplier    runtime.MemStats
}

// обновление метрик
func (ms *MetricsStorage) UpdateMetrics() {
	// считывание переменных их runtimr
	runtime.ReadMemStats(&ms.Supplier)
	// определение дополнительных метрик
	ms.PollCount += 1
	ms.RandomValue = rand.Float64()
}

// отправка метрик
func (ms *MetricsStorage) SendMetrics(IP string, port int) {
	metrics := make(map[string]any)
	metrics["Alloc"] = ms.Supplier.Alloc
	metrics["BuckHashSys"] = ms.Supplier.BuckHashSys
	metrics["Frees"] = ms.Supplier.Frees
	metrics["GCCPUFraction"] = ms.Supplier.GCCPUFraction
	metrics["GCSys"] = ms.Supplier.GCSys
	metrics["HeapAlloc"] = ms.Supplier.HeapAlloc
	metrics["HeapIdle"] = ms.Supplier.HeapIdle
	metrics["HeapInuse"] = ms.Supplier.HeapInuse
	metrics["HeapObjects"] = ms.Supplier.HeapObjects
	metrics["HeapReleased"] = ms.Supplier.HeapReleased
	metrics["HeapSys"] = ms.Supplier.HeapSys
	metrics["LastGC"] = ms.Supplier.LastGC
	metrics["Lookups"] = ms.Supplier.Lookups
	metrics["MCacheInuse"] = ms.Supplier.MCacheInuse
	metrics["MCacheSys"] = ms.Supplier.MCacheSys
	metrics["MSpanInuse"] = ms.Supplier.MSpanInuse
	metrics["MSpanSys"] = ms.Supplier.MSpanSys
	metrics["Mallocs"] = ms.Supplier.Mallocs
	metrics["NextGC"] = ms.Supplier.NextGC
	metrics["NumForcedGC"] = ms.Supplier.NumForcedGC
	metrics["NumGC"] = ms.Supplier.NumGC
	metrics["OtherSys"] = ms.Supplier.OtherSys
	metrics["PauseTotalNs"] = ms.Supplier.PauseTotalNs
	metrics["StackInuse"] = ms.Supplier.StackInuse
	metrics["StackSys"] = ms.Supplier.StackSys
	metrics["Sys"] = ms.Supplier.Sys
	metrics["TotalAlloc"] = ms.Supplier.TotalAlloc
	metrics["RandomValue"] = ms.RandomValue

	client := http.Client{}
	for key, value := range metrics {
		if err := sendToServer(client, IP, port, value, key); err != nil {
			log.Println(err)
		}
	}
	// отправка дополнительных параметров
	if err := sendToServer(client, IP, port, ms.PollCount, "PollCount"); err == nil {
		ms.PollCount = 0
	} else {
		log.Println(err)
	}
	log.Println("Metrics send iteration finished")
}

// отправка запроса к серверу
func sendToServer(client http.Client, IP string, port int, value any, name string) error {
	query := ""
	switch value.(type) {
	case uint32, int64, uint64:
		query = "counter"
	case float64:
		query = "gauge"
	default:
		return fmt.Errorf("metric '%s' type indefined: '%T'", name, value)
	}
	query = fmt.Sprintf("http://%s:%d/update/%s/%s/%v", IP, port, query, name, value)
	resp, err := client.Post(query, "text/plain", nil)
	if err != nil {
		return fmt.Errorf("send metric '%s' error: '%v'", name, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send metric '%s' statusCode error: %d", name, resp.StatusCode)
	}
	return nil
}
