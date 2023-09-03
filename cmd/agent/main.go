package main

import (
	"fmt"
	"log"
	"os"

	"github.com/gostuding/go-metrics/internal/agent"
	"go.uber.org/zap"
)

var (
	buildVersion = "N/A"
	buildDate    = "N/A"
	buildCommit  = "N/A"
)

func main() {
	fmt.Fprintf(os.Stdout, "Build version: %s\n", buildVersion)
	fmt.Fprintf(os.Stdout, "Build date: %s\n", buildDate)
	fmt.Fprintf(os.Stdout, "Build commit: %s\n", buildCommit)
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
