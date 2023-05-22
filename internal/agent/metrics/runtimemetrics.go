package metrics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"

	"github.com/dpanic/convert"
)

func NewMemoryStorage() *MetricsStorage {
	return &MetricsStorage{MetricsSlice: make(map[string]Metrics, 0)}
}

type MetricsStorage struct {
	RandomValue  float64
	PollCount    int64
	Supplier     runtime.MemStats
	MetricsSlice map[string]Metrics
}

type Metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func makeMetric(id string, value any) Metrics {
	switch value.(type) {
	case int, uint32, int64, uint64:
		val, _ := strconv.Atoi(fmt.Sprint(value))
		return Metrics{
			ID:    id,
			MType: "counter",
			Delta: convert.ToInt64P(val),
		}
	case float64:
		return Metrics{
			ID:    id,
			MType: "gauge",
			Value: convert.ToFloat64P(convert.ToFloat(value)),
		}
	}
	return Metrics{}
}

// обновление метрик
func (ms *MetricsStorage) UpdateMetrics() {
	// считывание переменных их runtimr
	var runtimeMetrics runtime.MemStats
	runtime.ReadMemStats(&runtimeMetrics)
	ms.MetricsSlice["Alloc"] = makeMetric("Alloc", runtimeMetrics.Alloc)
	ms.MetricsSlice["Frees"] = makeMetric("Frees", runtimeMetrics.Frees)
	ms.MetricsSlice["GCCPUFraction"] = makeMetric("GCCPUFraction", runtimeMetrics.GCCPUFraction)
	ms.MetricsSlice["GCSys"] = makeMetric("GCSys", runtimeMetrics.GCSys)
	ms.MetricsSlice["HeapAlloc"] = makeMetric("HeapAlloc", runtimeMetrics.HeapAlloc)
	ms.MetricsSlice["HeapIdle"] = makeMetric("HeapIdle", runtimeMetrics.HeapIdle)
	ms.MetricsSlice["HeapInuse"] = makeMetric("HeapInuse", runtimeMetrics.HeapInuse)
	ms.MetricsSlice["HeapObjects"] = makeMetric("HeapObjects", runtimeMetrics.HeapObjects)
	ms.MetricsSlice["HeapReleased"] = makeMetric("HeapReleased", runtimeMetrics.HeapReleased)
	ms.MetricsSlice["HeapSys"] = makeMetric("HeapSys", runtimeMetrics.HeapSys)
	ms.MetricsSlice["LastGC"] = makeMetric("LastGC", runtimeMetrics.LastGC)
	ms.MetricsSlice["Lookups"] = makeMetric("Lookups", runtimeMetrics.Lookups)
	ms.MetricsSlice["MCacheInuse"] = makeMetric("MCacheInuse", runtimeMetrics.MCacheInuse)
	ms.MetricsSlice["MCacheSys"] = makeMetric("MCacheSys", runtimeMetrics.MCacheSys)
	ms.MetricsSlice["MSpanInuse"] = makeMetric("MSpanInuse", runtimeMetrics.MSpanInuse)
	ms.MetricsSlice["MSpanSys"] = makeMetric("MSpanSys", runtimeMetrics.MSpanSys)
	ms.MetricsSlice["Mallocs"] = makeMetric("Mallocs", runtimeMetrics.Mallocs)
	ms.MetricsSlice["NextGC"] = makeMetric("NextGC", runtimeMetrics.NextGC)
	ms.MetricsSlice["NumForcedGC"] = makeMetric("NumForcedGC", runtimeMetrics.NumForcedGC)
	ms.MetricsSlice["NumGC"] = makeMetric("NumGC", runtimeMetrics.NumGC)
	ms.MetricsSlice["OtherSys"] = makeMetric("OtherSys", runtimeMetrics.OtherSys)
	ms.MetricsSlice["PauseTotalNs"] = makeMetric("PauseTotalNs", runtimeMetrics.PauseTotalNs)
	ms.MetricsSlice["StackInuse"] = makeMetric("StackInuse", runtimeMetrics.StackInuse)
	ms.MetricsSlice["StackSys"] = makeMetric("StackSys", runtimeMetrics.StackSys)
	ms.MetricsSlice["Sys"] = makeMetric("Sys", runtimeMetrics.Sys)
	ms.MetricsSlice["TotalAlloc"] = makeMetric("TotalAlloc", runtimeMetrics.TotalAlloc)
	if ms.MetricsSlice["PollCount"].Delta == nil {
		ms.MetricsSlice["PollCount"] = makeMetric("PollCount", 1)
	} else {
		ms.MetricsSlice["PollCount"] = makeMetric("PollCount", (*ms.MetricsSlice["PollCount"].Delta + 1))
	}
	ms.MetricsSlice["RandomValue"] = makeMetric("RandomValue", rand.Float64())
	log.Println("Update finished")
}

// отправка метрик
func (ms *MetricsStorage) SendMetrics(IP string, port int) {
	for _, metric := range ms.MetricsSlice {
		if err := sendJSONToServer(IP, port, metric); err != nil {
			log.Println(err)
		} else {
			if metric.ID == "PollCount" {
				ms.MetricsSlice["PollCount"] = makeMetric("PollCount", 0)
			}
		}
	}
	log.Println("Metrics send iteration finished")
}

// отправка запроса к серверу
func sendJSONToServer(IP string, port int, metric Metrics) error {
	client := http.Client{}
	query := fmt.Sprintf("http://%s:%d/update/", IP, port)
	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("metric convert error: %s", err)
	}
	buf := bytes.NewBuffer(body)
	resp, err := client.Post(query, "applications/json", buf)
	if err != nil {
		return fmt.Errorf("send metric '%s' error: '%v'", metric.ID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send metric '%v' statusCode error: %d", metric.ID, resp.StatusCode)
	}
	return nil
}
