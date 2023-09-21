package metrics

import (
	"fmt"
	"hash"
	"math/rand"
	"runtime"
	"strconv"
)

const (
	closeTimeout = 10
	counter      = "counter"
	gauge        = "gauge"
	pCount       = "PollCount"
)

// MakeMetric is private func for create metrics object from id:value values.
// It defines type of metrics from value's type (int64 or float64).
func makeMetric(id string, value any) (*metrics, error) {
	switch value.(type) {
	case int, uint32, int64, uint64:
		val, err := strconv.ParseInt(fmt.Sprint(value), 10, 64)
		if err != nil {
			return nil, fmt.Errorf("convert '%s' to int64 error: %w", id, err)
		}
		return &metrics{
			ID:    id,
			MType: counter,
			Delta: &val,
		}, nil
	case float64:
		val, err := strconv.ParseFloat(fmt.Sprint(value), 64)
		if err != nil {
			return nil, fmt.Errorf("convert '%s' to float64 error: %w", id, err)
		}
		return &metrics{
			ID:    id,
			MType: gauge,
			Value: &val,
		}, nil
	default:
		return nil, fmt.Errorf("convert error. metric '%s' type undefined", id)
	}
}

// MakeMap is private func for create metrics map[string]any from runtime.MemStats.
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
		mass[pCount] = 1
	} else {
		mass[pCount] = *pollCount + 1
	}
	return mass
}

// Internal function.
func hashToString(h hash.Hash) string {
	return fmt.Sprintf("%x", h.Sum(nil))
}
