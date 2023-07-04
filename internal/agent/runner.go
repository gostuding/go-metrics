package agent

import (
	"time"
)

// интерфейс для отправки и обновления данных
type Storager interface {
	UpdateMetrics()
	UpdateAditionalMetrics()
	SendMetricsSlice()
}

// бесконечный цикл отправки данных
func StartAgent(args Config, storage Storager) {
	pollTicker := time.NewTicker(time.Duration(args.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(args.ReportInterval) * time.Second)
	defer pollTicker.Stop()
	defer reportTicker.Stop()
	for {
		select {
		case <-pollTicker.C:
			go storage.UpdateMetrics()
			go storage.UpdateAditionalMetrics()
		case <-reportTicker.C:
			storage.SendMetricsSlice()
		}
	}
}
