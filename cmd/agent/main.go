package main

import (
	"github.com/gostuding/go-metrics/internal/agent"
	"github.com/gostuding/go-metrics/internal/agent/metrics"
)

// минимум кода в cmd
func main() {
	// процесс получения данных параметров можно изменить, т.е. считывать флаги из файла а не параметров
	address, updateTime, sendTime := agent.GetFlags() // определение входных параметров для отправки
	storage := &metrics.MetricsStorage{}              // определение источника метрик
	// запуск агента, логика его работы находится в internal
	agent.StartAgent(address.IP, address.Port, updateTime, sendTime, storage)
}
