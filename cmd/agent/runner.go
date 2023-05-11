package main

import (
	"fmt"
	"time"

	"github.com/gostuding/go-metrics/cmd/agent/metrics"
)

func DoAgent(storage metrics.Storager) {
	index := 1
	for {
		time.Sleep(1 * time.Second)
		if index%UpdateTime == 0 {
			storage.UpdateMetrics()
		}
		if index%SendTime == 0 {
			storage.SendMetrics(SendAddress)
			fmt.Println("Send")
		}
		index += 1
	}
}
