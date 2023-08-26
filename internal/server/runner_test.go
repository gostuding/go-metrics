package server

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/gostuding/go-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func createMemServer(ip string) (*Server, error) {
	logger, err := NewLogger()
	if err != nil {
		return nil, fmt.Errorf("logger create error: %w", err)
	}
	cfg, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("config create error: %w", err)
	}
	cfg.IPAddress = ip
	cfg.Restore = false
	storage, err := storage.NewMemStorage(cfg.Restore, cfg.FileStorePath, cfg.StoreInterval)
	if err != nil {
		return nil, fmt.Errorf("storage create error: %w", err)
	}
	return NewServer(cfg, logger, storage), nil
}

func getRandServerAddress() string {
	return fmt.Sprintf(":%d", (rand.Intn(1000) + 10000))
}

func TestServer_StopServer(t *testing.T) {
	runnedSrv, err := createMemServer(getRandServerAddress())
	if !assert.NoError(t, err, "create runner server error") {
		return
	}
	stoppedSrv, err := createMemServer(getRandServerAddress())
	if !assert.NoError(t, err, "create stopped server error") {
		return
	}
	go func() {
		err := runnedSrv.RunServer()
		assert.NoError(t, err, "runner server error")
	}()
	time.Sleep(time.Second)
	tests := []struct {
		name    string
		server  *Server
		wantErr bool
	}{
		{
			name:    "server not run",
			server:  stoppedSrv,
			wantErr: true,
		},
		{
			name:    "server stop success",
			server:  runnedSrv,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.server.StopServer(); (err != nil) != tt.wantErr {
				t.Errorf("StopServer() error = %v", err)
			}
			if tt.server.isRun {
				t.Error("StopServer() error. Server is run still")
			}
		})
	}
}

func TestServer_RunServer(t *testing.T) {
	srvForRun, err := createMemServer(getRandServerAddress())
	if !assert.NoError(t, err, "create srvForRun server error") {
		return
	}
	srvRunned, err := createMemServer(getRandServerAddress())
	if !assert.NoError(t, err, "create stoppedSrv server error") {
		return
	}
	go func() {
		err := srvRunned.RunServer()
		assert.NoError(t, err, "runner server error")
	}()

	tests := []struct {
		name    string
		server  *Server
		wantErr bool
	}{
		{
			name:    "success run",
			server:  srvForRun,
			wantErr: false,
		},
		{
			name:    "already run",
			server:  srvRunned,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			go func() {
				if err := tt.server.RunServer(); (err != nil) != tt.wantErr {
					t.Errorf("Server.RunServer() error = %v", err)
				}
			}()
			time.Sleep(time.Second)
			err := tt.server.StopServer()
			assert.NoError(t, err, "Stop server error")
		})
	}
}
