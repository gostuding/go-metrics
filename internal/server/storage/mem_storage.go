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

func (ms *memStorage) Update(ctx context.Context, mType string, mName string, mValue string) error {
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
func (ms *memStorage) GetMetric(ctx context.Context, mType string, mName string) (string, error) {
	switch mType {
	case "gauge":
		for key, val := range ms.Gauges {
			if key == mName {
				return fmt.Sprint(val), nil
			}
		}
	case "counter":
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
	case "counter":
		if m.Delta != nil {
			ms.Counters[m.ID] += *m.Delta
			delta := ms.Counters[m.ID]
			m.Delta = &delta
		} else {
			return nil, errors.New("metric's delta indefined")
		}
	case "gauge":
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
	item, err := ms.updateOneMetric(metric)
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

// очистка хранилища
func (ms *memStorage) Clear(ctx context.Context) error {
	for k := range ms.Counters {
		delete(ms.Counters, k)
	}
	for k := range ms.Gauges {
		delete(ms.Gauges, k)
	}
	ms.Gauges = make(map[string]float64)
	ms.Counters = make(map[string]int64)
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
	for index, value := range metrics {
		_, err := ms.updateOneMetric(value)
		if err != nil {
			resp += fmt.Sprintf("%d. '%s' update ERROR: %v\n", index+1, value.ID, err)
		} else {
			resp += fmt.Sprintf("%d. '%s' update SUCCESS \n", index+1, value.ID)
		}
	}
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
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return fmt.Errorf("database connect error: %v", err)
	}

	rows, err := db.Query("Select name, value from counters;")
	if err != nil {
		return fmt.Errorf("select counters error: %v", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("select counters error: %v", rows.Err())
	}
	for rows.Next() {
		var name string
		var value int64
		err := rows.Scan(&name, &value)
		if err != nil {
			return fmt.Errorf("scan counters values error: %v", err)
		}
		ms.Counters[name] = value
	}

	rows, err = db.Query("Select name, value from gauges;")
	if err != nil {
		return fmt.Errorf("select gauges error: %v", err)
	}
	if rows.Err() != nil {
		return fmt.Errorf("select gauges error: %v", rows.Err())
	}
	for rows.Next() {
		var name string
		var value float64
		err := rows.Scan(&name, &value)
		if err != nil {
			return fmt.Errorf("scan gauges values error: %v", err)
		}
		ms.Gauges[name] = value
	}

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
	err := checkDatabaseStructure(ms.ConnectDBString)
	if err != nil {
		return fmt.Errorf("check database structure error: %v", err)
	}
	return ms.fromDBRestore()
}

func (ms *memStorage) saveInFile() error {
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

func (ms *memStorage) saveInDatabase() error {
	db, err := sql.Open("pgx", ms.ConnectDBString)
	if err != nil {
		return fmt.Errorf("database connect error: %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	transaction, err := db.Begin()
	if err != nil {
		return fmt.Errorf("transaction create error: %v", err)
	}
	defer transaction.Rollback()

	_, err = transaction.ExecContext(ctx, "Delete from gauges")
	if err != nil {
		return fmt.Errorf("clear gauges table error: %v", err)
	}
	_, err = transaction.ExecContext(ctx, "Delete from counters")
	if err != nil {
		return fmt.Errorf("clear counters table error: %v", err)
	}
	for key, value := range ms.Gauges {
		_, err = transaction.ExecContext(ctx, "Insert into gauges (name, value) values($1, $2);", key, value)
		if err != nil {
			return fmt.Errorf("save gauge ('%s') value (%f) error: %v", key, value, err)
		}
	}
	for key, value := range ms.Counters {
		_, err = transaction.ExecContext(ctx, "Insert into counters (name, value) values($1, $2);", key, value)
		if err != nil {
			return fmt.Errorf("save gauge ('%s') value (%d) error: %v", key, value, err)
		}
	}
	err = transaction.Commit()
	if err != nil {
		return fmt.Errorf("transaction error: %v", err)
	}
	return nil
}

func (ms *memStorage) Save() error {
	if ms.ConnectDBString == "" {
		return ms.saveInFile()
	}
	return ms.saveInDatabase()
}
