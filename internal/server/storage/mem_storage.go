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

const (
	gaugeType    = "gauge"   // name for gauge type
	counterType  = "counter" // name for counter type
	fileOpenMode = 0644
	converError  = iota
	metricTypeIncorrect
	metricNotFoud
	metricTypeError
	saveMetricError
	jsonConverError
)

func makeError(errorType int, vals ...any) error {
	switch errorType {
	case converError:
		return fmt.Errorf("%s value convert error: %w", vals...)
	case metricTypeIncorrect:
		return errors.New("metric type incorrect. Availible types are: guage or counter")
	case metricNotFoud:
		return fmt.Errorf("metric '%s' with type '%s' not found", vals...)
	case metricTypeError:
		return errors.New("metric type error, use counter like int64 or gauge like float64")
	case saveMetricError:
		return fmt.Errorf("save metric error: %w", vals...)
	case jsonConverError:
		return fmt.Errorf("json conver error: %w", vals...)
	default:
		return fmt.Errorf("error type undefined: %d", errorType)
	}
}

type (
	// MemStorage contains metrics data in memory.
	MemStorage struct {
		Gauges       map[string]float64 `json:"gauges"`   // gauge metrics
		Counters     map[string]int64   `json:"counters"` // counter metrics
		SavePath     string             `json:"-"`        // path to file for save storage data
		SaveInterval int                `json:"-"`        // save data interval. If is 0 - storage saves in runtime.
		mx           sync.RWMutex       `json:"-"`        // mutex for storage
		Restore      bool               `json:"-"`        // flag for restore data from file
	}

	// Metric contains data about one metric.
	metric struct {
		Delta *int64   `json:"delta,omitempty"` // counter value
		Value *float64 `json:"value,omitempty"` // gauge value
		ID    string   `json:"id"`              // name
		MType string   `json:"type"`            // can be 'gauge' or 'counter'
	}
)

// NewMemStorage creates memStorage.
// If the restore flag is set, the data will be restored from the file,
// or the corresponding error will be returned.
func NewMemStorage(restore bool, filePath string, saveInterval int) (*MemStorage, error) {
	storage := MemStorage{
		Gauges:       make(map[string]float64),
		Counters:     make(map[string]int64),
		Restore:      restore,
		SavePath:     filePath,
		SaveInterval: saveInterval,
	}
	return &storage, storage.restore()
}

// Update creates or updates metric value in storage.
// Context doesn't have mean. Used to satisfy the interface.
func (ms *MemStorage) Update(
	ctx context.Context,
	mType string,
	mName string,
	mValue string,
) error {
	switch mType {
	case gaugeType:
		val, err := strconv.ParseFloat(mValue, 64)
		if err != nil {
			return makeError(converError, gaugeType, err)
		}
		ms.mx.Lock()
		ms.Gauges[mName] = val
		ms.mx.Unlock()
	case counterType:
		val, err := strconv.ParseInt(mValue, 10, 64)
		if err != nil {
			return makeError(converError, counterType, err)
		}
		ms.mx.Lock()
		ms.Counters[mName] += val
		ms.mx.Unlock()
	default:
		return makeError(metricTypeIncorrect)
	}
	if ms.SaveInterval == 0 {
		return ms.Save()
	}
	return nil
}

// GetMetric returns the metric value as string.
// Context doesn't have mean. Used to satisfy the interface.
func (ms *MemStorage) GetMetric(
	ctx context.Context,
	mType string,
	mName string,
) (string, error) {
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
				return strconv.FormatInt(val, 10), nil
			}
		}
	}
	return "", makeError(metricNotFoud, mName, mType)
}

// GetMetricsHTML returns all metrics values as HTML string.
// Context doesn't have mean. Used to satisfy the interface.
func (ms *MemStorage) GetMetricsHTML(ctx context.Context) (string, error) {
	gauges := make([]string, 0, len(ms.Gauges))
	counters := make([]string, 0, len(ms.Counters))
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	for _, key := range getSortedKeysFloat(ms.Gauges) {
		gauges = append(gauges, fmt.Sprintf("'%s'= %f", key, ms.Gauges[key]))
	}
	for _, key := range getSortedKeysInt(ms.Counters) {
		counters = append(counters, fmt.Sprintf("'%s'= %d", key, ms.Counters[key]))
	}

	return makeHTML(&gauges, &counters), nil
}

func makeHTML(gauges, counters *[]string) string {
	body := "<!doctype html> <html lang='en'> <head> <meta charset='utf-8'> <title>Список метрик</title></head>"
	body += "<body><header><h1><p>Metrics list</p></h1></header>"
	body += "<h1><p>Gauges</p></h1>"
	for index, value := range *gauges {
		body += makeMetricString(index+1, value)
	}
	body += "<h1><p>Counters</p></h1>"
	for index, value := range *counters {
		body += makeMetricString(index+1, value)
	}
	body += "</body></html>"
	return body
}

func makeMetricString(index int, value string) string {
	return fmt.Sprintf("<nav><p>%d. %s</p></nav>", index, value)
}

// UpdateOneMetric is private func for update storage.
func (ms *MemStorage) updateOneMetric(m metric) (*metric, error) {
	switch m.MType {
	case counterType:
		if m.Delta != nil {
			ms.Counters[m.ID] += *m.Delta
			delta := ms.Counters[m.ID]
			m.Delta = &delta
		} else {
			return nil, errors.New("delta indefined")
		}
	case gaugeType:
		if m.Value != nil {
			ms.Gauges[m.ID] = *m.Value
		} else {
			return nil, errors.New("value indefined")
		}
	default:
		return nil, makeError(metricTypeError)
	}
	return &m, nil
}

// UpdateJSON creates or updates metric value in storage.
// Gets []byte with JSON.
// Context doesn't have mean. Used to satisfy the interface.
func (ms *MemStorage) UpdateJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, fmt.Errorf("conver error: %w", err)
	}
	ms.mx.Lock()
	item, err := ms.updateOneMetric(metric)
	ms.mx.Unlock()
	if err != nil {
		return nil, err
	}
	resp, err := json.Marshal(item)
	if err != nil {
		return nil, fmt.Errorf("marshal to json error: %w", err)
	}

	if ms.SaveInterval == 0 {
		if err := ms.Save(); err != nil {
			return nil, makeError(saveMetricError, err)
		}
	}
	return resp, nil
}

// GetMetricJSON returns the metric value as string.
// Gets []byte with JSON.
// Context doesn't have mean. Used to satisfy the interface.
func (ms *MemStorage) GetMetricJSON(ctx context.Context, data []byte) ([]byte, error) {
	var metric metric
	err := json.Unmarshal(data, &metric)
	if err != nil {
		return nil, makeError(jsonConverError, err)
	}
	resp := make([]byte, 0)
	err = fmt.Errorf("metric not found. id: '%s', type: '%s'", metric.ID, metric.MType)
	ms.mx.RLock()
	defer ms.mx.RUnlock()
	switch metric.MType {
	case counterType:
		for key, val := range ms.Counters {
			val := val
			if key == metric.ID {
				metric.Delta = &val
				resp, err = json.Marshal(metric)
			}
		}
	case gaugeType:
		for key, val := range ms.Gauges {
			val := val
			if key == metric.ID {
				metric.Value = &val
				resp, err = json.Marshal(metric)
			}
		}
	default:
		return nil, fmt.Errorf("metric type ('%s') error", metric.MType)
	}
	if err != nil {
		return []byte(""), err
	}
	return resp, nil
}

// PingDB doesn't have mean. Used to satisfy the interface.
// Context doesn't have mean. Used to satisfy the interface.
// Always returns nil.
func (ms *MemStorage) PingDB(ctx context.Context) error {
	return nil
}

// Clear deletes all data from the storage.
// Context doesn't have mean. Used to satisfy the interface.
func (ms *MemStorage) Clear(ctx context.Context) error {
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

// UpdateJSONSlice updates the repository with metrics that are obtained
// by translating the received JSON into a list of metrics.
// Context doesn't have mean. Used to satisfy the interface.
func (ms *MemStorage) UpdateJSONSlice(ctx context.Context, data []byte) ([]byte, error) {
	var metrics []metric
	err := json.Unmarshal(data, &metrics)
	if err != nil {
		return nil, makeError(jsonConverError, err)
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
			return nil, makeError(saveMetricError, err)
		}
	}
	return []byte(resp), nil
}

// Save writes storage data to file.
func (ms *MemStorage) Save() error {
	if ms.SavePath == "" {
		return nil
	}
	file, err := os.OpenFile(ms.SavePath, os.O_WRONLY|os.O_TRUNC|os.O_CREATE, fileOpenMode)
	if err != nil {
		return fmt.Errorf("open file for save error: %w", err)
	}
	defer file.Close() //nolint:errcheck //<-senselessly
	ms.mx.Lock()
	data, err := json.MarshalIndent(ms, "", "    ")
	ms.mx.Unlock()
	if err != nil {
		return fmt.Errorf("marshal values error: %w", err)
	}
	_, err = file.Write(data)
	if err != nil {
		return fmt.Errorf("write file error: %w", err)
	}
	return nil
}

// GetSortedKeysFloat private func. Returns sorted list of map keys.
func getSortedKeysFloat(items map[string]float64) []string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// GetSortedKeysInt private func. Returns sorted list of map keys.
func getSortedKeysInt(items map[string]int64) []string {
	keys := make([]string, 0, len(items))
	for k := range items {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

// Restore is private func. Restores storage data from file.
func (ms *MemStorage) restore() error {
	if !ms.Restore {
		return nil
	}
	file, err := os.OpenFile(ms.SavePath, os.O_RDONLY|os.O_CREATE, fileOpenMode)
	if err != nil {
		return fmt.Errorf("open file error: %w", err)
	}
	decoder := json.NewDecoder(file)
	defer file.Close() //nolint:errcheck //<-senselessly
	err = decoder.Decode(ms)
	if err != nil && errors.Is(err, io.EOF) {
		return fmt.Errorf("decode error: %w", err)
	}
	return nil
}

// Stop saves data to file and clears storage.
func (ms *MemStorage) Stop() error {
	err := ms.Save()
	if err != nil {
		return fmt.Errorf("save storage duaring stop error: %w", err)
	}
	return nil
}
