package metrics

import (
	"fmt"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
)

// var GaugeType int = 1
// var ConterType int = 2

type MetricsStorage struct {
	RandomValue float64
	PollCount   int64
	Supplier    runtime.MemStats
}

func (ms *MetricsStorage) UpdateMetrics() {
	runtime.ReadMemStats(&ms.Supplier)
	ms.PollCount += 1
	ms.RandomValue = rand.Float64()
}

func (ms MetricsStorage) SendMetrics(address fmt.Stringer) {
	send := func(adr string, value any, name string) {
		query := ""
		switch value.(type) {
		case int64, uint64:
			query = "counter"
		case float64:
			query = "gauge"
		}
		if len(query) > 0 {
			query = fmt.Sprintf("http://%s/update/%s/%s/%v", adr, query, name, value)
			resp, err := http.Post(query, "text/plain", nil)
			if err != nil {
				fmt.Printf("Send metric: '%s' error: '%v'\n", name, err)
			}
			defer resp.Body.Close()
		}
	}

	fields := reflect.VisibleFields(reflect.TypeOf(ms.Supplier))
	for _, field := range fields {
		send(address.String(), reflect.ValueOf(ms.Supplier).FieldByName(field.Name).Interface(), field.Name)
	}
	send(address.String(), ms.PollCount, "PollCount")
	send(address.String(), ms.RandomValue, "RandomValue")
}
