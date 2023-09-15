// Package metrics is using for collect metrics and sending them to serer.
package metrics

import (
	"bytes"
	"compress/gzip"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"runtime"
	"sync"
	"time"

	"github.com/shirou/gopsutil/mem"
	"go.uber.org/zap"
)

// Const values.
const (
	hashVarName = "HashSHA256" // Header name for hash check.
)

type (
	// MetricsStorage is object for use as Storager interface.
	metricsStorage struct {
		URL          string             // URL for requests send to server
		MetricsSlice map[string]metrics // metrics storage
		PublicKey    *rsa.PublicKey     // encription messages key
		Logger       *zap.SugaredLogger // logger
		resiveChan   chan resiveStruct  // chan for read requests results
		requestChan  chan struct{}      // chan for make requests
		Key          []byte             // check hash key
		mx           sync.RWMutex       // mutex
		GzipCompress bool               // flag to use gzip compress
		Supplier     runtime.MemStats   // metrics data supplier
	}

	// Metrics is one metric struct.
	metrics struct {
		Value *float64 `json:"value,omitempty"` // gauge value
		Delta *int64   `json:"delta,omitempty"` // counter value
		ID    string   `json:"id"`              // metrics name
		MType string   `json:"type"`            // metrics type: gauge or counter
	}

	// ResiveStruct is internal struct.
	resiveStruct struct {
		Metric *metrics
		Err    error
	}
)

// NewMemoryStorage creates memory storage for metrics.
//
// Args:
// pk *rsa.PublicKey - public RSA key for messages encription
// logger *zap.Logger
// ip string - server ip address for send metrics
// key []byte - key for requests hash check
// port int - server port for send metrics
// compress bool - flag to compress data by gzip
// rateLimit int - max count requests in time.
func NewMemoryStorage(
	pk *rsa.PublicKey,
	logger *zap.Logger,
	ip string,
	key []byte,
	port int,
	compress bool,
	rateLimit int,
) *metricsStorage {
	mS := metricsStorage{
		MetricsSlice: make(map[string]metrics),
		Logger:       logger.Sugar(),
		PublicKey:    pk,
		GzipCompress: compress,
		Key:          key,
		URL:          fmt.Sprintf("http://%s/updates/", net.JoinHostPort(ip, fmt.Sprint(port))),
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

// AddMetric is private func and adds one metrics to MetricsSLice.
func (ms *metricsStorage) addMetric(name string, value any) {
	metric, err := makeMetric(name, value)
	if err != nil {
		ms.Logger.Warn(err)
	} else {
		ms.MetricsSlice[name] = *metric
	}
}

// UpdateAditionalMetrics collects metrics from mem.VirtualMemoryStat.
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

// UpdateMetrics collects metrics from runtime.MemStats.
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
}

// SendMetricsSlice sends metrics by JSON list.
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

// SendJSONToServer is private func for send requests to server.
func (ms *metricsStorage) sendJSONToServer(body []byte, metric *metrics) {
	defer func() {
		<-ms.requestChan
	}()
	client := http.Client{}
	req, err := http.NewRequest(http.MethodPost, ms.URL, nil)
	if err != nil {
		ms.resiveChan <- resiveStruct{Err: fmt.Errorf("request create error: %w", err), Metric: metric}
		return
	}
	if ms.PublicKey != nil {
		body, err = encriptMessage(body, ms.PublicKey)
		if err != nil {
			ms.Logger.Warnf("metrics encription error: %w", err)
			return
		}
	}
	if ms.GzipCompress {
		var b bytes.Buffer
		gz := gzip.NewWriter(&b)
		_, err = gz.Write(body)
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
		req.Header.Add(hashVarName, hashToString(h))
	}
	resp, err := client.Do(req)
	if err != nil {
		ms.resiveChan <- resiveStruct{Err: fmt.Errorf("send error: '%w'", err), Metric: metric}
		return
	}
	defer resp.Body.Close() //nolint:errcheck // <- senselessly
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
		if resp.Header.Get(hashVarName) != hashToString(hash) {
			ms.resiveChan <- resiveStruct{Err: errors.New("check responce hash summ error"), Metric: metric}
			return
		}
	}
	ms.resiveChan <- resiveStruct{Err: nil, Metric: metric}
}

// Close checks if the last data were send to server. If not, sends data to server.
func (ms *metricsStorage) Close() error {
	close(ms.resiveChan)
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(closeTimeout)*time.Second)
	defer cancel()
	ms.resiveChan = make(chan resiveStruct, 1)
	closeResive := true
	for _, m := range ms.MetricsSlice {
		if m.MType == counter && m.ID == pCount && m.Delta != nil && *m.Delta > 0 {
			ms.SendMetricsSlice()
			closeResive = false
		}
	}
	if closeResive {
		close(ms.resiveChan)
	}
	select {
	case r := <-ms.resiveChan:
		return r.Err
	case <-ctx.Done():
		return errors.New("close timeout error")
	}
}

// encription message.
func encriptMessage(msg []byte, key *rsa.PublicKey) ([]byte, error) {
	rng := rand.Reader
	hash := sha256.New()
	size := key.Size() - 2*hash.Size() - 2 //nolint:gomnd //<-default values
	encripted := make([]byte, 0)
	for _, slice := range splitMessage(msg, size) {
		data, err := rsa.EncryptOAEP(hash, rng, key, slice, []byte(""))
		if err != nil {
			return nil, fmt.Errorf("message encript error: %w", err)
		}
		encripted = append(encripted, data...)
	}
	fmt.Fprintln(os.Stdout, len(msg), len(encripted))
	return encripted, nil
}
