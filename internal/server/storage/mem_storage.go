package storage

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"sort"
	"strconv"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib"
)

// Структура для хранения данных о метриках
type memStorage struct {
	Gauges          map[string]float64 `json:"gauges"`
	Counters        map[string]int64   `json:"counters"`
	Restore         bool               `json:"-"`
	SavePath        string             `json:"-"`
	SaveInterval    int                `json:"-"`
	ConnectDBString string             `json:"-"`
}

// струтктура не экспоритуемая, т.к. сейчас это не нужно
type metric struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func NewMemStorage(restore bool, filePath string, saveInterval int, DBconnect string) (*memStorage, error) {
	storage := memStorage{
		Gauges:          make(map[string]float64),
		Counters:        make(map[string]int64),
		Restore:         restore,
		SavePath:        filePath,
		SaveInterval:    saveInterval,
		ConnectDBString: DBconnect,
	}
	return &storage, storage.restore()
}

func (ms *memStorage) Update(mType string, mName string, mValue string) error {
	switch mType {
	case "gauge":
		val, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return fmt.Errorf("gauge value convert error: %w", err)
		}
		ms.Gauges[mName] = val
	case "counter":
		val, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return fmt.Errorf("counter value convert error: %w", err)
		}
		ms.Counters[mName] += val
	default:
		return errors.New("metric type incorrect. Availible types are: guage or counter")
	}
	if ms.SaveInterval == 0 {
		ms.Save()
	}
	return nil
}

// Получение значения метрики по типу и имени
func (ms *memStorage) GetMetric(mType string, mName string) (string, error) {
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
	return "", fmt.Errorf("metric '%s' with type '%s' not found", mName, mType)
}

// Список всех метрик в html
func (ms *memStorage) GetMetricsHTML() string {
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

// обновление через json
func (ms *memStorage) UpdateJSON(data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
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
func (ms *memStorage) GetMetricJSON(data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("json conver error: %w", err)
	}
	resp := make([]byte, 0)
	err = fmt.Errorf("metric not found. id: '%s', type: '%s'", metric.ID, metric.MType)
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
		return []byte(""), err
	}
	return resp, nil
}

// проверка подключения к БД
func (ms *memStorage) PingDB(ctx context.Context) error {
	if ms.ConnectDBString == "" {
		return fmt.Errorf("connect DB string undefined")
	}
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return fmt.Errorf("database connect error: %w", err)
	}
	defer db.Close()
	ctx, cancel := context.WithTimeout(ctx, 1*time.Second)
	defer cancel()
	if err = db.PingContext(ctx); err != nil {
		return fmt.Errorf("check database ping error: %w", err)
	}
	return nil
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

// загрузка хранилища из файла
func (ms *memStorage) fromFileRestore() error {
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

// загрузка хранилища из БД
func (ms *memStorage) fromDBRestore() error {
	// проверка подключения к БД
	// для iter10 опускаем внутренности, т.к. нужно возвращать ошибки подключения
	return nil
}

// восстановление данных из хранилища
func (ms *memStorage) restore() error {
	if !ms.Restore {
		return nil
	}
	if ms.ConnectDBString == "" {
		return ms.fromFileRestore()
	}
	return ms.fromDBRestore()
}

func (ms *memStorage) Save() error {
	file, err := os.OpenFile(ms.SavePath, os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	data, err := json.MarshalIndent(ms, "", "    ")
	if err != nil {
		return err
	}
	_, err = file.Write(data)
	if err != nil {
		return err
	}
	return nil
}
