// Package agent is used to run the metrics collection agent.
// To start, use:
//
//	 cfg, err := agent.NewConfig()
//	 if err != nil {
//			log.Fatalln(err)
//	 }
//	 logger, err := zap.NewDevelopment()
//	 if err != nil {
//			log.Fatalln("create logger error:", err)
//	 }
//	 agent := NewAgent(cfg, logger)
//	 agent.StartAgent()
package agent
