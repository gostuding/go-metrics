package storage

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strconv"
)

// Структура для хранения данных о метриках
type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func (ms *MemStorage) Update(mType string, mName string, mValue string) (int, error) {
	if mType == "gauge" {
		val, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return http.StatusBadRequest, err
		}
		ms.addGauge(mName, val)
	} else if mType == "counter" {
		val, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return http.StatusBadRequest, err
		}
		ms.addCounter(mName, val)
	} else {
		log.Printf("Update metric error. Metric's type incorrect. Type is: %s\n", mType)
		return http.StatusBadRequest, errors.New("metric type incorrect. Availible types are: guage or counter")
	}
	return http.StatusOK, nil
}

// Получение значения метрики по типу и имени
func (ms *MemStorage) GetMetric(mType string, mName string) (string, int) {
	if mType == "gauge" {
		for key, val := range ms.Gauges {
			if key == mName {
				return fmt.Sprintf("%v", val), http.StatusOK
			}
		}
		fmt.Printf("Gauge metric not found by name: %s\n", mName)
	} else if mType == "counter" {
		for key, val := range ms.Counters {
			if key == mName {
				return fmt.Sprintf("%v", val), http.StatusOK
			}
		}
		fmt.Printf("Counter metric not found by name: %s\n", mName)
	} else {
		fmt.Printf("Get metric's type incorrect: %s\n", mType)
		return "", http.StatusNotFound
	}
	return "", http.StatusNotFound
}

// Список всех метрик в html
func (ms *MemStorage) GetMetricsHTML() string {
	body := "<!doctype html> <html lang='en'> <head> <meta charset='utf-8'> <title>Список метрик</title></head>"
	body += "<body><header><h1><p>Metrics list</p></h1></header>"
	index := 1
	body += "<h1><p>Gauges</p></h1>"
	for _, key := range getSortedxKeysFloat(ms.Gauges) {
		body += fmt.Sprintf("<nav><p>%d. '%s'= '%f'</p></nav>", index, key, ms.Gauges[key])
		index += 1
	}
	body += "<h1><p>Counters</p></h1>"
	index = 1
	for _, key := range getSortedKeysInt(ms.Counters) {
		body += fmt.Sprintf("<nav><p>%d. '%s'= '%d'</p></nav>", index, key, ms.Counters[key])
		index += 1
	}
	body += "</body></html>"
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

func getSortedxKeysFloat(items map[string]float64) []string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func getSortedKeysInt(items map[string]int64) []string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
