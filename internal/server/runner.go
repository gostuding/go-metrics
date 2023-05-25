package server

import (
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

var currentOptions ServerOptions
var done = make(chan bool)

// запуск сервера на прослушку
func RunServer(options *ServerOptions, storage Storage) error {
	err := InitLogger()
	if err != nil {
		return fmt.Errorf("create logger error: %v", err)
	}
	if options == nil {
		return fmt.Errorf("server options error")
	}
	currentOptions = *options
	Logger.Infoln("Run server at adress: ", options.IpAddress)
	go saveStorageInterval(storage)
	go saveStorageBeforeFinish(storage)
	err = http.ListenAndServe(options.IpAddress, makeRouter(storage))
	if err != nil {
		return fmt.Errorf("run server error: %v", err)
	}
	done <- true
	return nil
}

func saveStorageInterval(storage Storage) {
	if currentOptions.StoreInterval < 1 {
		return
	}
	ticker := time.NewTicker(time.Duration(currentOptions.StoreInterval) * time.Second)
	Logger.Infof("save storage interval: %d sec.", currentOptions.StoreInterval)
	defer ticker.Stop()
	for {
		select {
		case <-done:
			Logger.Info("stop save storage interval")
			return
		case <-ticker.C:
			err := storage.Save()
			if err != nil {
				Logger.Warnf("save storage error: %v\n", err)
			} else {
				Logger.Info("save storage by interval")
			}
		}
	}
}

func saveStorageBeforeFinish(storage Storage) {
	signalChanel := make(chan os.Signal, 1)
	signal.Notify(signalChanel,
		syscall.SIGHUP,
		syscall.SIGINT,
		syscall.SIGTERM,
		syscall.SIGQUIT)
	exit_chan := make(chan int)
	go func() {
		for {
			s := <-signalChanel
			switch s {
			case syscall.SIGHUP, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
				err := storage.Save()
				if err == nil {
					Logger.Info("save storage before finish")
				} else {
					Logger.Warnln("save storage error", err)
				}
				exit_chan <- 0
			default:
				Logger.Warnln("undefined system signal")
				exit_chan <- 1
			}
		}
	}()
	os.Exit(<-exit_chan)
}
