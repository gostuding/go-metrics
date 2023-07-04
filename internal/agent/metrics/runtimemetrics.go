package metrics

import (
	"bytes"
	"compress/gzip"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"runtime"
	"strconv"
	"sync"

	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

type metricsStorage struct {
	Supplier     runtime.MemStats
	MetricsSlice map[string]metrics
	Logger       *zap.SugaredLogger
	IP           string
	Port         int
	GzipCompress bool
	Key          []byte
	URL          string
	mx           sync.RWMutex
	resiveChan   chan resiveStruct
	requestChan  chan struct{}
}

type metrics struct {
	ID    string   `json:"id"`              // имя метрики
	MType string   `json:"type"`            // параметр, принимающий значение gauge или counter
	Delta *int64   `json:"delta,omitempty"` // значение метрики в случае передачи counter
	Value *float64 `json:"value,omitempty"` // значение метрики в случае передачи gauge
}

type resiveStruct struct {
	Metric *metrics
	Err    error
}

func NewMemoryStorage(logger *zap.Logger, ip string, key []byte, port int, compress bool, rateLimit int) *metricsStorage {
	mS := metricsStorage{
		MetricsSlice: make(map[string]metrics),
		Logger:       logger.Sugar(),
		IP:           ip,
		Port:         port,
		GzipCompress: compress,
		Key:          key,
		URL:          fmt.Sprintf("http://%s:%d/updates/", ip, port),
		resiveChan:   make(chan resiveStruct, rateLimit),
		requestChan:  make(chan struct{}, rateLimit),
	}

	go func() {
		for item := range mS.resiveChan {
			if item.Err != nil {
				mS.Logger.Warnf("send error: %w", item.Err)
			} else {
				mS.mx.Lock()
				if item.Metric.ID == "PollCount" {
					delta := *mS.MetricsSlice["PollCount"].Delta - *item.Metric.Delta
					mS.MetricsSlice["PollCount"] = metrics{ID: "PollCount", MType: "counter", Delta: &delta}
				}
				mS.mx.Unlock()
			}
		}
	}()
	return &mS
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

func (ms *metricsStorage) addMetric(name string, value any) {
	metric, err := makeMetric(name, value)
	if err != nil {
		ms.Logger.Warn(err)
	} else {
		ms.MetricsSlice[name] = *metric
	}
}

func (ms *metricsStorage) UpdateAditionalMetrics() {
	memory, err := mem.VirtualMemory()
	if err != nil {
		ms.Logger.Warnf("get virtualmemory metric error: %w", err)
		return
	}

	mSlice := make(map[string]float64)
	mSlice["TotalMemory"] = float64(memory.Total)
	mSlice["FreeMemory"] = float64(memory.Free)
	mSlice["UsedMemoryPercent"] = memory.UsedPercent
	mSlice["CPUutilization1"] = float64(runtime.NumCPU())

	ms.mx.Lock()
	for name, value := range mSlice {
		ms.addMetric(name, value)
	}
	ms.mx.Unlock()
}

func (ms *metricsStorage) UpdateMetrics() {
	var rStats runtime.MemStats
	runtime.ReadMemStats(&rStats)
	ms.mx.Lock()
	for name, value := range makeMap(&rStats, ms.MetricsSlice["PollCount"].Delta) {
		ms.addMetric(name, value)
		if ms.MetricsSlice[name].MType == "counter" {
			ms.addMetric(fmt.Sprintf("%sGauge", name), float64(*ms.MetricsSlice[name].Delta))
		}
	}
	ms.mx.Unlock()
	ms.Logger.Debugln("Update finished")
}

func (ms *metricsStorage) SendMetricsSlice() {
	mSlice := make([]metrics, 0)

	ms.mx.RLock()
	defer ms.mx.RUnlock()
	for _, item := range ms.MetricsSlice {
		mSlice = append(mSlice, item)
	}

	body, err := json.Marshal(mSlice)
	if err != nil {
		ms.Logger.Warnf("metrics slice conver error: %w", err)
		return
	}
	select {
	case ms.requestChan <- struct{}{}:
		metric := ms.MetricsSlice["PollCount"]
		go ms.sendJSONToServer(body, &metric)
		ms.Logger.Debug("Metrics slice send success")
	default:
		ms.Logger.Warnln("send metric slice error. Chan is full.")
	}
}

func (ms *metricsStorage) sendJSONToServer(body []byte, metric *metrics) {
	defer func() {
		<-ms.requestChan
	}()

	client := http.Client{}
	req, err := http.NewRequest("POST", ms.URL, nil)
	if err != nil {
		ms.resiveChan <- resiveStruct{Err: fmt.Errorf("request create error: %w", err), Metric: metric}
		return
	}
	if ms.GzipCompress {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		_, err := gz.Write(body)
		if err != nil {
			ms.resiveChan <- resiveStruct{Err: fmt.Errorf("compress error: %w", err), Metric: metric}
			return
		}
		err = gz.Close()
		if err != nil {
			ms.resiveChan <- resiveStruct{Err: fmt.Errorf("compressor close error: %w", err), Metric: metric}
			return
		}
		body = b.Bytes()
		req.Header.Add("Content-Encoding", "gzip")
	}
	req.Body = io.NopCloser(bytes.NewReader(body))
	if ms.Key != nil {
		h := hmac.New(sha256.New, ms.Key)
		_, err = h.Write(body)
		if err != nil {
			ms.resiveChan <- resiveStruct{Err: fmt.Errorf("write hash summ error: '%w'", err), Metric: metric}
			return
		}
		req.Header.Add("HashSHA256", fmt.Sprintf("%x", h.Sum(nil)))
	}
	resp, err := client.Do(req)
	if err != nil {
		ms.resiveChan <- resiveStruct{Err: fmt.Errorf("send error: '%w'", err), Metric: metric}
		return
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		ms.resiveChan <- resiveStruct{Err: fmt.Errorf("statusCode error: %d", resp.StatusCode), Metric: metric}
		return
	}
	if ms.Key != nil {
		data, err := io.ReadAll(resp.Body)
		if err != nil {
			ms.resiveChan <- resiveStruct{Err: fmt.Errorf("responce body read error: %w", err), Metric: metric}
			return
		}
		hash := hmac.New(sha256.New, ms.Key)
		_, err = hash.Write(data)
		if err != nil {
			ms.resiveChan <- resiveStruct{Err: fmt.Errorf("responce read hash summ error: '%w'", err), Metric: metric}
			return
		}
		if resp.Header.Get("HashSHA256") != fmt.Sprintf("%x", hash.Sum(nil)) {
			ms.resiveChan <- resiveStruct{Err: errors.New("check responce hash summ error"), Metric: metric}
			return
		}
	}
	ms.resiveChan <- resiveStruct{Err: nil, Metric: metric}
}
