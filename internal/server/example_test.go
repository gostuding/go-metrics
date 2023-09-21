package server

import (
	"fmt"
	"log"
	"time"

	"github.com/gostuding/go-metrics/internal/server/storage"
)

func ExampleNewServer() {
	logger, err := NewLogger()
	if err != nil {
		log.Fatalf("logger create error: %v", err)
	}
	cfg := &Config{IPAddress: defaultAddress, StoreInterval: saveInterval}
	storage, err := storage.NewMemStorage(restoreStorage, defFileName, saveInterval)
	if err != nil {
		logger.Warnf("storage create error: %w", err)
		return
	}
	srv := NewServer(cfg, logger, storage)
	fmt.Printf("Server default address: '%s'", srv.Config.IPAddress)

	// Output:
	// Server default address: ':8080'
}

func ExampleNewConfig() {
	cfg, err := NewConfig()
	if err != nil {
		fmt.Printf("create config error: %v", err)
	} else {
		fmt.Printf("New config with address: %s", cfg.IPAddress)
	}

	// Output:
	// New config with address: :8080
}

func ExampleNewLogger() {
	logger, err := NewLogger()
	if err != nil {
		fmt.Printf("create logger error: %v", err)
	} else {
		logger.Debug("Logger created")
	}

	// Output:
	//
}

func ExampleServer_RunServer() {
	logger, err := NewLogger()
	if err != nil {
		log.Fatalf("logger create error: %v", err)
	}
	cfg, err := NewConfig()
	if err != nil {
		log.Fatalf("config create error: %v", err)
	}
	storage, err := storage.NewMemStorage(cfg.Restore, cfg.FileStorePath, cfg.StoreInterval)
	if err != nil {
		logger.Warnf("storage create error: %w", err)
		return
	}
	srv := NewServer(cfg, logger, storage)
	defer srv.StopServer() //nolint:errcheck //<-senselessly
	// Run server in gorutine for not block main thread
	go func() {
		fmt.Println("Run server success")
		if err = srv.RunServer(); err != nil {
			fmt.Printf("Run error: %v", err)
			logger.Warnf("Run server errro: %w", err)
		}
	}()
	time.Sleep(time.Second)

	// Output:
	// Run server success
}

func ExampleServer_StopServer() {
	// Create server.
	srv, err := createMemServer(defaultAddress)
	if err != nil {
		fmt.Printf("create server error: %v", err)
		return
	}
	// Run server in other gorutine.
	go func() {
		if err = srv.RunServer(); err != nil {
			fmt.Printf("Run server errro: %v", err)
		}
	}()

	// ...
	// Do anything.
	time.Sleep(time.Second)
	// ...

	// Stop server work
	err = srv.StopServer()
	if err != nil {
		fmt.Printf("stop server error: %v", err)
	} else {
		fmt.Println("Stop server success")
	}

	// Output:
	// Stop server success
}
