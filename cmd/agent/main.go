package main

import (
	"log"

	"github.com/gostuding/go-metrics/internal/agent"
	"go.uber.org/zap"
)

func main() {
	cfg, err := agent.NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln("create logger error:", err)
	}
	agent := agent.NewAgent(cfg, logger)
	agent.StartAgent()
}
