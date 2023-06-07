package main

import (
	"log"

	"github.com/gostuding/go-metrics/internal/agent"
	"github.com/gostuding/go-metrics/internal/agent/metrics"
	"go.uber.org/zap"
)

func main() {
	agentArgs, err := agent.GetFlags()
	if err != nil {
		log.Fatalln(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln("create logger error:", err)
	}
	storage := metrics.NewMemoryStorage(logger)
	agent.StartAgent(agentArgs, storage)
}
