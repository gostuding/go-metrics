package server

import (
	"fmt"

	"go.uber.org/zap"
)

// InitLogger is create new logger with Suger type.
func InitLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("logger init error: %w", err)
	}
	return logger.Sugar(), nil
}
