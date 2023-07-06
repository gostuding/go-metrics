package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go.uber.org/zap"
)

// запуск сервера на прослушку
func RunServer(options *Config, storage Storage, logger *zap.SugaredLogger) error {
	if options == nil {
		return fmt.Errorf("server options error")
	}
	logger.Infoln("Run server at adress: ", options.IPAddress)
	if options.ConnectDBString == "" {
		go saveStorageInterval(options.StoreInterval, storage, logger)
		go saveStorageBeforeFinish(storage, logger)
	}
	return http.ListenAndServe(options.IPAddress, makeRouter(storage, logger, options.Key))
}

func saveStorageInterval(interval int, storage Storage, logger *zap.SugaredLogger) {
	if interval < 1 {
		logger.Infoln("save storage runtime mode", interval)
		return
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	logger.Infof("save storage interval: %d sec.", interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		err := storage.Save()
		if err != nil {
			logger.Warnf("save storage error: %w", err)
		} else {
			logger.Info("save storage by interval")
		}
	}
}

func saveStorageBeforeFinish(storage Storage, logger *zap.SugaredLogger) {
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signalChanel
	err := storage.Save()
	if err == nil {
		logger.Info("save storage before finish")
	} else {
		logger.Warnln("save storage in finish error", err)
	}
	os.Exit(0)
}
