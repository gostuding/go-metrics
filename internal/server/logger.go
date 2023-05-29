package server

import (
	"time"

	"go.uber.org/zap"
)

// определение объекта для логирования. До инициализации - не выводит сообщений
var Logger *zap.SugaredLogger = zap.NewNop().Sugar()

// TODO обёртка для вывода типизированного лога о запросах
// в дальнейще либо убрать, либо переделать для универсального использования
func requestLog(URI string, method string, duration time.Duration) {
	Logger.Infow(
		"Server logger",
		"type", "request",
		"uri", URI,
		"method", method,
		"duration", duration,
	)
}

func responseLog(URI string, status int, size int) {
	Logger.Infow(
		"Server logger",
		"type", "responce",
		zap.String("uri", URI),
		zap.Int("status", status),
		zap.Int("size", size),
	)
}

// инициализация логера и определение его типа как Sugar
func InitLogger() error {
	logger, err := zap.NewDevelopment()
	if err != nil {
		return err
	}
	Logger = logger.Sugar()
	return nil
}
