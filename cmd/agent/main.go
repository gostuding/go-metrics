package main

import (
	"github.com/gostuding/go-metrics/internal/agent"
	"github.com/gostuding/go-metrics/internal/agent/metrics"
)

func main() {
	agentArgs, err := agent.GetFlags()
	if err != nil {
		panic(err)
	}
	storage, err := metrics.NewMemoryStorage()
	if err != nil {
		panic(err)
	}
	agent.StartAgent(agentArgs, storage)
}
