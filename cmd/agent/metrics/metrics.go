package metrics

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
)

var GaugeType int = 1
var ConterType int = 2

type MetricsStorage struct {
	RandomValue float64
	PollCount   int64
	Supplier    runtime.MemStats
}

func NewMetricStorage() *MetricsStorage {
	return new(MetricsStorage)
}

func (ms *MetricsStorage) UpdateMetrics() {
	runtime.ReadMemStats(&ms.Supplier)
	ms.PollCount += 1
	ms.RandomValue = rand.Float64()
}

func (ms MetricsStorage) SendMetrics() {
	send := func(value any, name string) {
		query := ""
		switch value.(type) {
		case int64, uint64:
			query = "counter"
		case float64:
			query = "gauge"
		}
		if len(query) > 0 {
			query = fmt.Sprintf("http://localhost:8080/update/%s/%s/%v", query, name, value)
			resp, err := http.Post(query, "text/plain", nil)
			if err != nil {
				fmt.Printf("Send metric: '%s' error: '%v'\n", name, err)
			} else {
				defer resp.Body.Close()
			}
		}
	}

	fields := reflect.VisibleFields(reflect.TypeOf(ms.Supplier))
	for _, field := range fields {
		send(reflect.ValueOf(ms.Supplier).FieldByName(field.Name).Interface(), field.Name)
	}
	send(ms.PollCount, "PollCount")
	send(ms.RandomValue, "RandomValue")
}
