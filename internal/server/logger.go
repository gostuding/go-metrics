package server

import (
	"go.uber.org/zap"
)

// определение объекта для логирования. До инициализации - не выводит сообщений
var Logger *zap.SugaredLogger = zap.NewNop().Sugar()

// инициализация логера и определение его типа как Sugar
func InitLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	Logger = logger.Sugar()
	return nil
}
