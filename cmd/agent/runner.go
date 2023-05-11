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
		if index%UpdateTime == 0 { // переменная из flags.go
			storage.UpdateMetrics()
		}
		if index%SendTime == 0 { // переменная из flags.go
			storage.SendMetrics(SendAddress) // переменная из flags.go
			fmt.Println("Send")
		}
		index += 1
	}
}
