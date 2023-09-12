package agent

import (
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/gostuding/go-metrics/internal/agent/metrics"
	"go.uber.org/zap"
)

// Struct for send data to server.
type (
	Agent struct {
		cfg      *Config // configuration
		logger   *zap.Logger
		Storage  Storager // storage for agent.
		stopChan chan os.Signal
		mutex    sync.Mutex
		isRun    bool
	}

	// Storager interface for metrics collecting.
	Storager interface {
		UpdateMetrics()
		UpdateAditionalMetrics()
		SendMetricsSlice()
	}
)

// NewAgent creates new Agent object.
func NewAgent(cfg *Config, logger *zap.Logger) *Agent {
	s := metrics.NewMemoryStorage(cfg.PublicKey, logger, cfg.IP, cfg.Key, cfg.Port, cfg.GzipCompress, cfg.RateLimit)
	return &Agent{Storage: s, logger: logger, cfg: cfg}
}

// StartAgent starts gorutines for update and send metrics.
func (a *Agent) StartAgent() {
	a.mutex.Lock()
	if a.isRun {
		a.mutex.Unlock()
		return
	}
	a.isRun = true
	a.stopChan = make(chan os.Signal, 1)
	signal.Notify(a.stopChan, os.Interrupt)
	a.mutex.Unlock()
	a.logger.Debug("Start agent")
	pollTicker := time.NewTicker(time.Duration(a.cfg.PollInterval) * time.Second)
	reportTicker := time.NewTicker(time.Duration(a.cfg.ReportInterval) * time.Second)
	defer pollTicker.Stop()
	defer reportTicker.Stop()
	for {
		select {
		case <-pollTicker.C:
			go a.Storage.UpdateMetrics()
			go a.Storage.UpdateAditionalMetrics()
		case <-reportTicker.C:
			a.Storage.SendMetricsSlice()
		case <-a.stopChan:
			a.logger.Debug("Agent work finished")
			return
		}
	}
}

// StopAgent finishing agent if it was start.
func (a *Agent) StopAgent() {
	a.logger.Debug("Stop agent")
	a.mutex.Lock()
	defer a.mutex.Unlock()
	if !a.isRun {
		return
	}
	a.isRun = false
	close(a.stopChan)
}

// IsRun - flag to show is agent run.
func (a *Agent) IsRun() bool {
	return a.isRun
}
