package server

import (
	"fmt"
	"math/rand"
	"testing"
	"time"

	"github.com/gostuding/go-metrics/internal/server/storage"
	"github.com/stretchr/testify/assert"
)

func createMemServer(IP string) (*Server, error) {
	logger, err := NewLogger()
	if err != nil {
		return nil, fmt.Errorf("logger create error: %v", err)
	}
	cfg, err := NewConfig()
	if err != nil {
		return nil, fmt.Errorf("config create error: %v", err)
	}
	cfg.IPAddress = IP
	storage, err := storage.NewMemStorage(cfg.Restore, cfg.FileStorePath, cfg.StoreInterval)
	if err != nil {
		return nil, fmt.Errorf("storage create error: %w", err)
	}
	return NewServer(cfg, logger, storage), nil
}

func getRandServerAddress() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf(":%d", (rand.Intn(1000) + 10000))
}

func TestServer_StopServer(t *testing.T) {
	runnedSrv, err := createMemServer(":8080")
	assert.NoError(t, err, "create runner server error")
	stoppedSrv, err := createMemServer(getRandServerAddress())
	assert.NoError(t, err, "create stopped server error")
	go runnedSrv.RunServer()
	time.Sleep(time.Second)
	tests := []struct {
		name    string
		server  *Server
		wantErr bool
	}{
		{
			name:    "server not run",
			wantErr: true,
			server:  stoppedSrv,
		},
		{
			name:    "server stop success",
			wantErr: false,
			server:  runnedSrv,
		},
	}
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.server.StopServer(); (err != nil) != tt.wantErr {
				t.Errorf("StopServer() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.server.isRun {
				t.Error("StopServer() error. Server is run still")
			}
		})
	}
}

func TestServer_RunServer(t *testing.T) {
	srvForRun, err := createMemServer(getRandServerAddress())
	assert.NoError(t, err, "create srvForRun server error")

	srvRunned, err := createMemServer(getRandServerAddress())
	assert.NoError(t, err, "create srvRunned server error")
	go srvRunned.RunServer()

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
					t.Errorf("Server.RunServer() error = %v, wantErr %v", err, tt.wantErr)
				}
			}()
			time.Sleep(time.Second)
			tt.server.StopServer() //nolint:wrapcheck // <- senselessly
		})
	}
}
