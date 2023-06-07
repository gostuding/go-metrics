package server

import (
	"fmt"

	"go.uber.org/zap"
)

// инициализация логера и определение его типа как Sugar
func InitLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("logger init error: %w", err)
	}
	return logger.Sugar(), nil
}
