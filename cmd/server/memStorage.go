package main

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
)

// Структура для хранения данных о метриках
type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

// создание нового объекта MemStorage
func NewMemStorage() *MemStorage {
	return new(MemStorage)
}

func (ms *MemStorage) AddMetric(path string) (int, error) {
	items := strings.Split(path, "/")
	if len(items) >= 5 {
		items = items[2:5]
	}
	if len(items) != 3 {
		fmt.Printf("Metric parse error. Len not equal to 3. %v\n", items)
		return http.StatusNotFound, errors.New("metric parse error. Checks url path and repeat")
	}
	if items[0] == "gauge" {
		val, err := strconv.ParseFloat(items[2], 64)
		if err != nil {
			return http.StatusBadRequest, err
		}
		ms.addGauge(items[1], val)
	} else if items[0] == "counter" {
		val, err := strconv.ParseInt(items[2], 10, 64)
		if err != nil {
			return http.StatusBadRequest, err
		}
		ms.addCounter(items[1], val)
	} else {
		fmt.Printf("Metric's type incorrect. Type is: %s\n", items[0])
		return http.StatusBadRequest, errors.New("metric type incorrect. Availible types are: guage or counter")
	}
	return http.StatusOK, nil
}

// Функция для удовлетворения интерфейсу Stringer
func (ms MemStorage) String() string {
	index := 1
	body := "==== MemoryStorage ====\n"
	for key, value := range ms.Gauges {
		body += fmt.Sprintf("Gauges: %d, name: '%s', value: '%f'\n", index, key, value)
		index += 1
	}
	index = 1
	for key, value := range ms.Counters {
		body += fmt.Sprintf("Counter: %d, name: '%s', value: '%d'\n", index, key, value)
		index += 1
	}
	body += "======================="
	return body
}

func (ms *MemStorage) addGauge(name string, value float64) {
	if ms.Gauges == nil {
		ms.Gauges = make(map[string]float64)
	}
	ms.Gauges[name] = value
}

func (ms *MemStorage) addCounter(name string, value int64) {
	if ms.Counters == nil {
		ms.Counters = make(map[string]int64)
	}
	ms.Counters[name] += value
}
