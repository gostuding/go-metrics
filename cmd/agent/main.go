package main

import (
	"log"

	"github.com/gostuding/go-metrics/internal/agent"
	"github.com/gostuding/go-metrics/internal/agent/metrics"
)

func main() {
	agentArgs, err := agent.GetFlags()
	if err != nil {
		log.Fatalf("run arguments incorret: %v", err)
	}
	storage := metrics.NewMemoryStorage()
	agent.StartAgent(agentArgs, storage)
}
