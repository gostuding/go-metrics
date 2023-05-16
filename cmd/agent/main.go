package main

import (
	"log"

	"github.com/gostuding/go-metrics/internal/agent"
	"github.com/gostuding/go-metrics/internal/agent/metrics"
)

func main() {
	agentArgs, err := agent.GetFlags()
	if err != nil {
		log.Fatal("run arguments incorret: ", err)
	}
	storage := &metrics.MetricsStorage{}
	agent.StartAgent(agentArgs, storage)
}
