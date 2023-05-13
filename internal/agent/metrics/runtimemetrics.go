package metrics

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime"
)

type MetricsStorage struct {
	RandomValue float64
	PollCount   int64
	Supplier    runtime.MemStats
}

// обновление данных
func (ms *MetricsStorage) UpdateMetrics() {
	// считывание переменных их runtimr
	runtime.ReadMemStats(&ms.Supplier)
	// определение дополнительных метрик
	ms.PollCount += 1
	ms.RandomValue = rand.Float64()
}

// отправка даннх
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
	// выборка всех переменных из пакета runtime
	fields := reflect.VisibleFields(reflect.TypeOf(ms.Supplier))
	for _, field := range fields {
		if err := send(client, reflect.ValueOf(ms.Supplier).FieldByName(field.Name).Interface(), field.Name); err != nil {
			log.Print(err)
		}
	}
	// отправка дополнительных параметров
	if err := send(client, ms.PollCount, "PollCount"); err == nil { // проверка на отправку данных для обнуления счётчика обновления метрики
		ms.PollCount = 0 // обнуляем счётчик для корректного отображения количества на сервере
	} else {
		log.Print(err)
	}
	if err := send(client, ms.RandomValue, "RandomValue"); err != nil {
		log.Print(err)
	}
}
