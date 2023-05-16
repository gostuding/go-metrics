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
	send := func(client http.Client, value any, name string) error {
		query := ""
		switch value.(type) {
		case int64, uint64:
			query = "counter"
		case float64:
			query = "gauge"
		default:
			return nil // тип переменной не подходит для отправки
		}
		query = fmt.Sprintf("http://%s:%d/update/%s/%s/%v", IP, port, query, name, value)
		resp, err := client.Post(query, "text/plain", nil)
		if err != nil {
			return fmt.Errorf("send metric '%s' error: '%v'", name, err)
		}
		defer resp.Body.Close()
		return nil
	}
	client := http.Client{}
	if err := send(client, ms.Supplier.Alloc, "Alloc"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.BuckHashSys, "BuckHashSys"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.Frees, "Frees"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.GCCPUFraction, "GCCPUFraction"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.GCSys, "GCSys"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.HeapAlloc, "HeapAlloc"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.HeapIdle, "HeapIdle"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.HeapInuse, "HeapInuse"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.HeapObjects, "HeapObjects"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.HeapReleased, "HeapReleased"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.HeapSys, "HeapSys"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.LastGC, "LastGC"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.Lookups, "Lookups"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.MCacheInuse, "MCacheInuse"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.MCacheSys, "MCacheSys"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.MCacheSys, "MCacheSys"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.MSpanInuse, "MSpanInuse"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.MSpanSys, "MSpanSys"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.Mallocs, "Mallocs"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.NextGC, "NextGC"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.NumForcedGC, "NumForcedGC"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.NumGC, "NumGC"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.OtherSys, "OtherSys"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.PauseTotalNs, "PauseTotalNs"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.StackInuse, "StackInuse"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.StackSys, "StackSys"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.Sys, "Sys"); err != nil {
		log.Println(err)
	}
	if err := send(client, ms.Supplier.TotalAlloc, "TotalAlloc"); err != nil {
		log.Println(err)
	}
	// отправка дополнительных параметров
	if err := send(client, ms.PollCount, "PollCount"); err == nil {
		ms.PollCount = 0
	} else {
		log.Println(err)
	}
	if err := send(client, ms.RandomValue, "RandomValue"); err != nil {
		log.Println(err)
	}
	log.Println("Metrics send iteration finished")
}
