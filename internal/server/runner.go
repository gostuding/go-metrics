package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// запуск сервера на прослушку
func RunServer(options *Config, storage Storage) error {
	err := InitLogger()
	if err != nil {
		return fmt.Errorf("create logger error: %v", err)
	}
	if options == nil {
		return fmt.Errorf("server options error")
	}
	Logger.Infoln("Run server at adress: ", options.IPAddress)
	go saveStorageInterval(options.StoreInterval, storage)
	go saveStorageBeforeFinish(storage)
	return http.ListenAndServe(options.IPAddress, makeRouter(storage))
}

func saveStorageInterval(interval int, storage Storage) {
	if interval < 1 {
		Logger.Infoln("save storage runtime mode", interval)
		return
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	Logger.Infof("save storage interval: %d sec.", interval)
	defer ticker.Stop()
	for {
		<-ticker.C
		err := storage.Save()
		if err != nil {
			Logger.Warnf("save storage error: %w", err)
		} else {
			Logger.Info("save storage by interval")
		}
	}
}

func saveStorageBeforeFinish(storage Storage) {
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel, syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-signalChanel
	err := storage.Save()
	if err == nil {
		Logger.Info("save storage before finish")
	} else {
		Logger.Warnln("save storage in finish error:", err)
	}
	os.Exit(0)
}
