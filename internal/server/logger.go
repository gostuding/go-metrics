package server

import (
	"fmt"

	"go.uber.org/zap"
)

// NewLogger is create new logger with Suger type.
func NewLogger() (*zap.SugaredLogger, error) {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return nil, fmt.Errorf("logger init error: %w", err)
	}
	return logger.Sugar(), nil
}
