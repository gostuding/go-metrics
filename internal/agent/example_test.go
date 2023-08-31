package agent

import (
	"fmt"
	"log"
	"time"

	"go.uber.org/zap"
)

func Example() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln("create logger error:", err)
	}
	agent := NewAgent(cfg, logger)
	// Run agent gorutine for not block main thread
	go agent.StartAgent()
	// Do any actions ...
	// ...
	time.Sleep(time.Second)

	// Stop agent.
	agent.StopAgent()

	// Check agent is run.
	if agent.IsRun() {
		fmt.Println("Agent is steel run")
	} else {
		fmt.Println("Agent stopped")
	}

	// Output:
	// Agent stopped
}

func ExampleNewConfig() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	fmt.Println(cfg)

	// Output:
	// :8080 -r 2 -p 10
}

func ExampleNewAgent() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln("create logger error:", err)
	}
	agent := NewAgent(cfg, logger)
	fmt.Println(agent.IsRun())

	// Output:
	// false
}

func ExampleAgent_StartAgent() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln("create logger error:", err)
	}
	agent := NewAgent(cfg, logger)
	go agent.StartAgent()

	time.Sleep(time.Second)
	fmt.Println(agent.IsRun())

	// Output:
	// true
}

func ExampleAgent_StopAgent() {
	cfg, err := NewConfig()
	if err != nil {
		log.Fatalln(err)
	}
	logger, err := zap.NewDevelopment()
	if err != nil {
		log.Fatalln("create logger error:", err)
	}
	agent := NewAgent(cfg, logger)
	go agent.StartAgent()
	// Do anything ...
	time.Sleep(time.Second)
	// Finish agent work.
	agent.StopAgent()
	fmt.Println(agent.IsRun())

	// Output:
	// false
}
