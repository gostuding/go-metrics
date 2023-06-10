package agent

import (
	"time"
)

// интерфейс для отправки и обновления данных
type Storager interface {
	UpdateMetrics()
	SendMetrics(string, int, bool)
	SendMetricsSlice(string, int, bool)
}

// бесконечный цикл отправки данных
func StartAgent(args Config, storage Storager) {
	pollTicker := time.NewTicker(time.Duration(args.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(args.ReportInterval) * time.Second)
	reportSliceTicker := time.NewTicker(time.Duration(args.ReportSliceInterval) * time.Second)
	defer pollTicker.Stop()
	defer reportTicker.Stop()
	defer reportSliceTicker.Stop()
	for {
		select {
		case <-pollTicker.C:
			storage.UpdateMetrics()
		case <-reportTicker.C:
			storage.SendMetrics(args.IP, args.Port, args.GzipCompress)
		case <-reportSliceTicker.C:
			storage.SendMetricsSlice(args.IP, args.Port, args.GzipCompress)
		}
	}
}
