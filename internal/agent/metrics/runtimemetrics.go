package metrics

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"

	"go.uber.org/zap"
)

type metricsStorage struct {
	Supplier     runtime.MemStats
	MetricsSlice map[string]metrics
	Logger       *zap.SugaredLogger
}

func NewMemoryStorage(logger *zap.Logger) *metricsStorage {
	return &metricsStorage{
		MetricsSlice: make(map[string]metrics),
		Logger:       logger.Sugar(),
	}
}

type metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

func makeMetric(id string, value any) (*metrics, error) {
	switch value.(type) {
	case int, uint32, int64, uint64:
		val, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("convert '%s' to int64 error: %w", id, err)
		}
		return &metrics{
			ID:    id,
			MType: "counter",
			Delta: &val,
		}, nil
	case float64:
		val, err := strconv.ParseFloat(fmt.Sprint(value), 64)
		if err != nil {
			return nil, fmt.Errorf("convert '%s' to float64 error: %w", id, err)
		}
		return &metrics{
			ID:    id,
			MType: "gauge",
			Value: &val,
		}, nil
	default:
		return nil, fmt.Errorf("convert error. metric '%s' type undefined", id)
	}
}

func makeMap(r *runtime.MemStats, pollCount *int64) map[string]any {
	mass := make(map[string]any)
	mass["Alloc"] = r.Alloc
	mass["BuckHashSys"] = r.BuckHashSys
	mass["Frees"] = r.Frees
	mass["GCCPUFraction"] = r.GCCPUFraction
	mass["GCSys"] = r.GCSys
	mass["HeapAlloc"] = r.HeapAlloc
	mass["HeapIdle"] = r.HeapIdle
	mass["HeapInuse"] = r.HeapInuse
	mass["HeapObjects"] = r.HeapObjects
	mass["HeapReleased"] = r.HeapReleased
	mass["HeapSys"] = r.HeapSys
	mass["LastGC"] = r.LastGC
	mass["Lookups"] = r.Lookups
	mass["MCacheInuse"] = r.MCacheInuse
	mass["MCacheSys"] = r.MCacheSys
	mass["MSpanInuse"] = r.MSpanInuse
	mass["MSpanSys"] = r.MSpanSys
	mass["Mallocs"] = r.Mallocs
	mass["NextGC"] = r.NextGC
	mass["NumForcedGC"] = r.NumForcedGC
	mass["NumGC"] = r.NumGC
	mass["OtherSys"] = r.OtherSys
	mass["PauseTotalNs"] = r.PauseTotalNs
	mass["StackInuse"] = r.StackInuse
	mass["StackSys"] = r.StackSys
	mass["TotalAlloc"] = r.TotalAlloc
	mass["Sys"] = r.Sys
	mass["RandomValue"] = rand.Float64()
	if pollCount == nil {
		mass["PollCount"] = 1
	} else {
		mass["PollCount"] = *pollCount + 1
	}
	return mass
}

// обновление метрик
func (ms *metricsStorage) UpdateMetrics() {
	var rStats runtime.MemStats
	runtime.ReadMemStats(&rStats)
	for name, value := range makeMap(&rStats, ms.MetricsSlice["PollCount"].Delta) {
		metric, err := makeMetric(name, value)
		if err != nil {
			ms.Logger.Warn(err)
			continue
		}
		ms.MetricsSlice[name] = *metric
		if metric.MType == "counter" {
			gaugeName := fmt.Sprintf("%sGauge", name)
			gauge, err := makeMetric(name, float64(*metric.Delta))
			if err != nil {
				ms.Logger.Warnf("make gauge value error: %w", err)
				continue
			}
			ms.MetricsSlice[gaugeName] = *gauge
		}
	}
	ms.Logger.Debugln("Update finished")
}

// отправка метрик
func (ms *metricsStorage) SendMetrics(IP string, port int, gzipCompress bool) {
	for _, metric := range ms.MetricsSlice {
		if err := sendJSONToServer(IP, port, metric, gzipCompress); err != nil {
			ms.Logger.Warn(err)
			continue
		}
		if metric.ID == "PollCount" && metric.MType == "counter" {
			ms.MetricsSlice["PollCount"] = metrics{ID: "PollCount", MType: "counter"}
		}
	}
	ms.Logger.Debugln("Metrics send iteration finished")
}

// отправка запроса к серверу
func sendJSONToServer(IP string, port int, metric metrics, compress bool) error {
	body, err := json.Marshal(metric)
	if err != nil {
		return fmt.Errorf("metric convert error: %s", err)
	}
	client := http.Client{}
	if compress {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		_, err = gz.Write(body)
		if err != nil {
			return fmt.Errorf("compress error: %w", err)
		}
		err = gz.Close()
		if err != nil {
			return fmt.Errorf("compressor close error: %w", err)
		}
		body = b.Bytes()
	}
	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/update/", IP, port), bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("request create error: %w", err)
	}
	if compress {
		req.Header.Add("Content-Encoding", "gzip")
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("send metric '%s' error: '%w'", metric.ID, err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("send metric '%v' statusCode error: %d", metric.ID, resp.StatusCode)
	}
	return nil
}

// отправка метрик списком
func (ms *metricsStorage) SendMetricsSlice(IP string, port int, gzipCompress bool) {
	mSlice := make([]metrics, 0)
	for _, item := range ms.MetricsSlice {
		mSlice = append(mSlice, item)
	}
	body, err := json.Marshal(mSlice)
	if err != nil {
		ms.Logger.Warnf("metrics slice conver error: %w", err)
		return
	}
	client := http.Client{}
	if gzipCompress {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		_, err = gz.Write(body)
		if err != nil {
			ms.Logger.Warnf("compress metrics json error: %w", err)
			return
		}
		err = gz.Close()
		if err != nil {
			ms.Logger.Warnf("compressor close error: %w", err)
			return
		}
		body = b.Bytes()
	}

	req, err := http.NewRequest("POST", fmt.Sprintf("http://%s:%d/updates/", IP, port), bytes.NewReader(body))
	if err != nil {
		ms.Logger.Warnf("request create error: %w", err)
		return
	}
	if gzipCompress {
		req.Header.Add("Content-Encoding", "gzip")
	}

	resp, err := client.Do(req)
	if err != nil {
		ms.Logger.Warnf("send metrics slice error: '%w'", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		ms.Logger.Warnf("send metrics slice statusCode error: %d", resp.StatusCode)
		return
	}

	for _, metric := range ms.MetricsSlice {
		if metric.ID == "PollCount" && metric.MType == "counter" {
			delta := int64(0)
			ms.MetricsSlice["PollCount"] = metrics{ID: "PollCount", MType: "counter", Delta: &delta}
		}
	}
	ms.Logger.Debugln("Metrics send iteration finished")
}
