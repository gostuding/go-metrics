package server

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	shutdownTimeout = 10 // timeout to stop server
)

// Server is struct for object.
type Server struct {
	Config  *Config            // server's options
	Storage Storage            // Storage interface
	Logger  *zap.SugaredLogger // server's logger
	isRun   bool               // flag to check is server run
	srv     http.Server        // internal server
	mutex   sync.Mutex
}

// Storage is interface for work with storage.
type Storage interface {
	StorageSetter
	StorageGetter
	StorageDB
}

// NewServer creates new server object.
func NewServer(config *Config, logger *zap.SugaredLogger, storage Storage) *Server {
	return &Server{Config: config, Logger: logger, Storage: storage}
}

// RunServer func runrs server. If the storage type is memory,
// runs too gorutines for save storage data by interval and
// save storage before finish work.
func (s *Server) RunServer() error {
	if s.isRun {
		return fmt.Errorf("server already run")
	}
	if s.Config == nil {
		return fmt.Errorf("server options is nil")
	}
	if s.Logger == nil {
		return fmt.Errorf("server logger is nil")
	}
	if s.Storage == nil {
		return fmt.Errorf("server storage is nil")
	}
	s.Logger.Infoln("Run server at adress: ", s.Config.IPAddress)
	ctx, cancelFunc := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancelFunc()
	srvChan := make(chan error, 1)
	s.srv = http.Server{Addr: s.Config.IPAddress, Handler: makeRouter(s.Storage, s.Logger, s.Config.Key)}
	go s.startServe(srvChan)
	s.mutex.Lock()
	s.isRun = true
	s.mutex.Unlock()
	go func() {
		<-ctx.Done()
		if err := s.StopServer(); err != nil {
			s.Logger.Warnf("stop server error: %w", err)
		}
	}()
	if s.Config.ConnectDBString == "" {
		go saveStorageInterval(ctx, s.Config.StoreInterval, s.Storage, s.Logger)
	}
	return <-srvChan
}

// StopServer is used for correct finish server's work.
func (s *Server) StopServer() error {
	if !s.isRun {
		return fmt.Errorf("the server is not running yet")
	}
	shtCtx, cancelFunc := context.WithTimeout(context.Background(), time.Duration(shutdownTimeout)*time.Second)
	defer cancelFunc()
	if err := s.srv.Shutdown(shtCtx); err != nil {
		return fmt.Errorf("shutdown server erorr: %w", err)
	}
	s.mutex.Lock()
	s.isRun = false
	s.mutex.Unlock()
	return nil
}

// startServe is private function for listen server's address and write error in chan when server finished.
func (s *Server) startServe(srvChan chan error) {
	err := s.srv.ListenAndServe()
	if err := s.Storage.Stop(); err != nil {
		s.Logger.Warnf("stop storage error: %w", err)
	} else {
		s.Logger.Debugln("Storage finished")
	}
	if errors.Is(err, http.ErrServerClosed) {
		srvChan <- nil
	} else {
		s.Logger.Warnf("server listen error: %w", err)
		srvChan <- err
	}
	s.Logger.Debugln("Server listen finished")
	close(srvChan)
}

// saveStorageInterval is private gorutine for save memory storage data by interval.
func saveStorageInterval(
	ctx context.Context,
	interval int,
	storage Storage,
	logger *zap.SugaredLogger,
) {
	if interval < 1 {
		logger.Infoln("save storage runtime mode", interval)
		return
	}
	ticker := time.NewTicker(time.Duration(interval) * time.Second)
	logger.Infof("save storage interval: %d sec.", interval)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			logger.Debugln("Save interval finished")
			return
		case <-ticker.C:
			err := storage.Save()
			if err != nil {
				logger.Warnf("save storage error: %w", err)
			} else {
				logger.Info("save storage by interval")
			}
		}
	}
}
