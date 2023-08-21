package server

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/gostuding/go-metrics/internal/server/storage"
)

var (
	defFileName    = filepath.Join(os.TempDir(), "Memory.strg")
	restoreStorage = false
	saveInterval   = 300
)

func Example() {
	logger, err := NewLogger()
	if err != nil {
		log.Fatalf("logger create error: %v", err)
	}
	cfg, err := NewConfig()
	if err != nil {
		logger.Warnf("config create error: %w", err)
		return
	}
	storage, err := storage.NewMemStorage(restoreStorage, defFileName, saveInterval)
	if err != nil {
		logger.Warnf("storage create error: %w", err)
		return
	}
	srv := NewServer(cfg, logger, storage)
	// Run server in gorutine for not block main thread
	go func() {
		if err = srv.RunServer(); err != nil {
			logger.Warnf("Run server errro: %w", err)
		}
	}()
	time.Sleep(time.Second)
	// Stop server
	err = srv.StopServer()
	if err != nil {
		fmt.Printf("stop server error: %v", err)
	}
}
