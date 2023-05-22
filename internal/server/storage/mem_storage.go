package storage

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strconv"
)

// Структура для хранения данных о метриках
// #TODO - сделать структуру не экспортируемой
type MemStorage struct {
	Gauges   map[string]float64
	Counters map[string]int64
}

// струтктура не экспоритуемая, т.к. сейчас это не нужно
type metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewMemStorage() *MemStorage {
	return &MemStorage{Gauges: make(map[string]float64), Counters: make(map[string]int64)}
}

func (ms *MemStorage) Update(mType string, mName string, mValue string) error {
	switch mType {
	case "gauge":
		val, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return err
		}
		ms.Gauges[mName] = val
	case "counter":
		val, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return err
		}
		ms.Counters[mName] += val
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
	for _, key := range getSortedKeysFloat(ms.Gauges) {
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

func getSortedKeysFloat(items map[string]float64) []string {
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

// обновление через json
func (ms *MemStorage) UpdateJSON(data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %s", err)
	}
	switch metric.MType {
	case "counter":
		if metric.Delta != nil {
			ms.Counters[metric.ID] += *metric.Delta
			delta := ms.Counters[metric.ID]
			metric.Delta = &delta
		} else {
			return nil, errors.New("metric's delta indefined")
		}
	case "gauge":
		if metric.Value != nil {
			ms.Gauges[metric.ID] += *metric.Value
		} else {
			return nil, errors.New("metric's value indefined")
		}
	default:
		return nil, errors.New("metric type error, use counter like int64 or gauge like float64")
	}
	resp, err := json.Marshal(metric)
	if err != nil {
		return nil, fmt.Errorf("convert to json error: %s", err)
	}
	return resp, nil
}

// обновление через json
func (ms *MemStorage) GetMetricJSON(data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %s", err)
	}
	resp := make([]byte, 0)
	err = errors.New("metric undefined")
	switch metric.MType {
	case "counter":
		for key, val := range ms.Counters {
			if key == metric.ID {
				metric.Delta = &val
				resp, err = json.Marshal(metric)
			}
		}
	case "gauge":
		for key, val := range ms.Gauges {
			if key == metric.ID {
				metric.Value = &val
				resp, err = json.Marshal(metric)
			}
		}
	default:
		return nil, fmt.Errorf("metric type ('%s') error, use counter like int64 or gauge like float64", metric.MType)
	}
	if err != nil {
		return nil, fmt.Errorf("convert to json error: %s", err)
	}
	return resp, nil
}
