package storage

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"sync"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var (
	gaugeType   = "gauge"
	counterType = "counter"
)

// Структура для хранения данных о метриках
type memStorage struct {
	Gauges       map[string]float64 `json:"gauges"`
	Counters     map[string]int64   `json:"counters"`
	Restore      bool               `json:"-"`
	SavePath     string             `json:"-"`
	SaveInterval int                `json:"-"`
	mx           sync.RWMutex       `json:"-"`
}

// струтктура не экспоритуемая, т.к. сейчас это не нужно
type metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewMemStorage(restore bool, filePath string, saveInterval int) (*memStorage, error) {
	storage := memStorage{
		Gauges:       make(map[string]float64),
		Counters:     make(map[string]int64),
		Restore:      restore,
		SavePath:     filePath,
		SaveInterval: saveInterval,
	}
	return &storage, storage.restore()
}

func (ms *memStorage) Update(ctx context.Context, mType string, mName string, mValue string) error {
	switch mType {
	case gaugeType:
		val, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return fmt.Errorf("gauge value convert error: %w", err)
		}
		ms.mx.Lock()
		ms.Gauges[mName] = val
		ms.mx.Unlock()
	case counterType:
		val, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return fmt.Errorf("counter value convert error: %w", err)
		}
		ms.mx.Lock()
		ms.Counters[mName] += val
		ms.mx.Unlock()
	default:
		return errors.New("metric type incorrect. Availible types are: guage or counter")
	}
	if ms.SaveInterval == 0 {
		ms.Save()
	}
	return nil
}

// Получение значения метрики по типу и имени
func (ms *memStorage) GetMetric(ctx context.Context, mType string, mName string) (string, error) {
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	switch mType {
	case gaugeType:
		for key, val := range ms.Gauges {
			if key == mName {
				return strconv.FormatFloat(val, 'f', -1, 64), nil
			}
		}
	case counterType:
		for key, val := range ms.Counters {
			if key == mName {
				return fmt.Sprintf("%d", val), nil
			}
		}
	}
	return "", fmt.Errorf("metric '%s' with type '%s' not found", mName, mType)
}

// Список всех метрик в html
func (ms *memStorage) GetMetricsHTML(ctx context.Context) (string, error) {
	body := "<!doctype html> <html lang='en'> <head> <meta charset='utf-8'> <title>Список метрик</title></head>"
	body += "<body><header><h1><p>Metrics list</p></h1></header>"
	index := 1
	body += "<h1><p>Gauges</p></h1>"
	ms.mx.RLock()
	defer ms.mx.RUnlock()
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
	return body, nil
}

func (ms *memStorage) updateOneMetric(m metric) (*metric, error) {
	switch m.MType {
	case counterType:
		if m.Delta != nil {
			ms.Counters[m.ID] += *m.Delta
			delta := ms.Counters[m.ID]
			m.Delta = &delta
		} else {
			return nil, errors.New("metric's delta indefined")
		}
	case gaugeType:
		if m.Value != nil {
			ms.Gauges[m.ID] = *m.Value
		} else {
			return nil, errors.New("metric's value indefined")
		}
	default:
		return nil, errors.New("metric type error, use counter like int64 or gauge like float64")
	}
	return &m, nil
}

// обновление через json
func (ms *memStorage) UpdateJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}
	ms.mx.Lock()
	item, err := ms.updateOneMetric(metric)
	ms.mx.Unlock()
	if err != nil {
		return nil, err
	}
	resp, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("convert to json error: %w", err)
	}

	if ms.SaveInterval == 0 {
		if err := ms.Save(); err != nil {
			return nil, fmt.Errorf("save metric error: %w", err)
		}
	}
	return resp, nil
}

// запрос метрик через json
func (ms *memStorage) GetMetricJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}
	resp := make([]byte, 0)
	err = fmt.Errorf("metric not found. id: '%s', type: '%s'", metric.ID, metric.MType)
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	switch metric.MType {
	case counterType:
		for key, val := range ms.Counters {
			if key == metric.ID {
				metric.Delta = &val
				resp, err = json.Marshal(metric)
			}
		}
	case gaugeType:
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
		return []byte(""), err
	}
	return resp, nil
}

// проверка подключения к БД
func (ms *memStorage) PingDB(ctx context.Context) error {
	return nil
}

// очистка хранилища
func (ms *memStorage) Clear(ctx context.Context) error {
	ms.mx.Lock()
	for k := range ms.Counters {
		delete(ms.Counters, k)
	}
	for k := range ms.Gauges {
		delete(ms.Gauges, k)
	}
	ms.Gauges = make(map[string]float64)
	ms.Counters = make(map[string]int64)
	ms.mx.Unlock()
	return ms.Save()
}

// обновление через json slice
func (ms *memStorage) UpdateJSONSlice(ctx context.Context, data []byte) ([]byte, error) {
	var metrics []metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}
	resp := ""
	ms.mx.Lock()
	for index, value := range metrics {
		_, err := ms.updateOneMetric(value)
		if err != nil {
			resp += fmt.Sprintf("%d. '%s' update ERROR: %v\n", index+1, value.ID, err)
		} else {
			resp += fmt.Sprintf("%d. '%s' update SUCCESS \n", index+1, value.ID)
		}
	}
	ms.mx.Unlock()
	if ms.SaveInterval == 0 {
		if err := ms.Save(); err != nil {
			return nil, fmt.Errorf("save metric error: %w", err)
		}
	}
	return []byte(resp), nil
}

// -------------------------------------------------------------------------------------------------
// внутренние функции хранилища
// -------------------------------------------------------------------------------------------------
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

// восстановление данных из хранилища
func (ms *memStorage) restore() error {
	if !ms.Restore {
		return nil
	}
	file, err := os.OpenFile(ms.SavePath, os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	decoder := json.NewDecoder(file)
	defer file.Close()
	err = decoder.Decode(ms)
	if err != nil && err != io.EOF {
		return err
	}
	return nil
}

func (ms *memStorage) Save() error {
	if ms.SavePath == "" {
		return nil
	}
	file, err := os.OpenFile(ms.SavePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	ms.mx.Lock()
	data, err := json.MarshalIndent(ms, "", "    ")
	ms.mx.Unlock()
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
