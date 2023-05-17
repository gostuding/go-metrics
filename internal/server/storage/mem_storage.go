package storage

import (
	"errors"
	"fmt"
	"sort"
	"strconv"
)

// Структура для хранения данных о метриках
type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

func (ms *MemStorage) Update(mType string, mName string, mValue string) error {
	switch mType {
	case "gauge":
		val, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return err
		}
		ms.addGauge(mName, val)
	case "counter":
		val, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return err
		}
		ms.addCounter(mName, val)
	default:
		return errors.New("metric type incorrect. Availible types are: guage or counter")
	}
	return nil
}

// Получение значения метрики по типу и имени
func (ms *MemStorage) GetMetric(mType string, mName string) (string, error) {
	switch mType {
	case "gauge":
		for key, val := range ms.Gauges {
			if key == mName {
				return fmt.Sprintf("%v", val), nil
			}
		}
	case "counter":
		for key, val := range ms.Counters {
			if key == mName {
				return fmt.Sprintf("%v", val), nil
			}
		}
	}
	return "", fmt.Errorf("metrick '%s' with type '%s' not found", mName, mType)
}

// Список всех метрик в html
func (ms *MemStorage) GetMetricsHTML() string {
	body := "<!doctype html> <html lang='en'> <head> <meta charset='utf-8'> <title>Список метрик</title></head>"
	body += "<body><header><h1><p>Metrics list</p></h1></header>"
	index := 1
	body += "<h1><p>Gauges</p></h1>"
	for _, key := range getSortedxKeysFloat(ms.Gauges) {
		body += fmt.Sprintf("<nav><p>%d. '%s'= %f</p></nav>", index, key, ms.Gauges[key])
		index += 1
	}
	body += "<h1><p>Counters</p></h1>"
	index = 1
	for _, key := range getSortedKeysInt(ms.Counters) {
		body += fmt.Sprintf("<nav><p>%d. '%s'= %d</p></nav>", index, key, ms.Counters[key])
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
