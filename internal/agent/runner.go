package agent

import (
	"time"
)

// интерфейс для отправки и обновления данных
type Storager interface {
	UpdateMetrics()
	UpdateAditionalMetrics()
	SendMetrics()
	SendMetricsSlice()
	// SetChans(*time.Ticker, *time.Ticker, *time.Ticker)
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
			go storage.UpdateMetrics()
			go storage.UpdateAditionalMetrics()
		case <-reportTicker.C:
			go storage.SendMetrics()
		case <-reportSliceTicker.C:
			go storage.SendMetricsSlice()
		}
	}
}
