package server

import (
	"context"
	"io"
	"net/http"

	"go.uber.org/zap"
)

// -----------------------------------------------------------------------------------
// Определение интерфейсов
// -----------------------------------------------------------------------------------
// Интерфей для установки значений в объект из строки
type StorageSetter interface {
	Update(context.Context, string, string, string) error
	UpdateJSON(context.Context, []byte) ([]byte, error)
	UpdateJSONSlice(context.Context, []byte) ([]byte, error)
	Save() error
}

// Интерфейс получения значения метрики
type StorageGetter interface {
	GetMetric(context.Context, string, string) (string, error)
	GetMetricJSON(context.Context, []byte) ([]byte, error)
}

// интерфейс для работы с БД
type StorageDB interface {
	PingDB(context.Context) error
	Clear(context.Context) error
}

// Интерфейс для вывод значений в виде HTML
type HTMLGetter interface {
	GetMetricsHTML(context.Context) (string, error)
}

// -----------------------------------------------------------------------------------
// Определение функций, которые используют интерфейсы
// -----------------------------------------------------------------------------------
type getMetricsArgs struct {
	mType string
	mName string
}

type updateMetricsArgs struct {
	base   getMetricsArgs
	mValue string
}

// Обработка запроса на добавление или изменение метрики
func Update(writer http.ResponseWriter, request *http.Request, storage StorageSetter, metric updateMetricsArgs, logger *zap.SugaredLogger) {
	if err := storage.Update(request.Context(), metric.base.mType, metric.base.mName, metric.mValue); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("update metric error: %w", err)
		return
	}
	writer.WriteHeader(http.StatusOK)
	logger.Debugf("update metric '%s' success", metric.base.mType)
}

// Обработка запроса значения метрики
func GetMetric(writer http.ResponseWriter, request *http.Request, storage StorageGetter, metric getMetricsArgs, logger *zap.SugaredLogger) {
	value, err := storage.GetMetric(request.Context(), metric.mType, metric.mName)
	if err != nil {
		writer.WriteHeader(http.StatusNotFound)
		logger.Warn(err)
		return
	}
	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write([]byte(value))
	if err != nil {
		logger.Warnf("write data to client error: %w", err)
	}
}

// Запрос всех метрик в html
func GetAllMetrics(writer http.ResponseWriter, request *http.Request, storage HTMLGetter, logger *zap.SugaredLogger) {
	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(http.StatusOK)
	data, err := storage.GetMetricsHTML(request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("get metrics in html error: %w", err)
		return
	}
	_, err = writer.Write([]byte(data))
	if err != nil {
		logger.Warnf("write metrics data to client error: %w", err)
	}
}

// обновление в JSON формате
func UpdateJSON(writer http.ResponseWriter, request *http.Request, storage StorageSetter, logger *zap.SugaredLogger) {
	writer.Header().Set("Content-Type", "application/json")
	data, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("read request body error: %w", err)
		return
	}
	value, err := storage.UpdateJSON(request.Context(), data)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("update metric error: %w", err)
		return
	}
	writer.WriteHeader(http.StatusOK)
	logger.Debug("update metric by json success")
	_, err = writer.Write(value)
	if err != nil {
		logger.Warnf("write data to clie`nt error: %w", err)
	}
}

// получение метрики в JSON формате
func GetMetricJSON(writer http.ResponseWriter, request *http.Request, storage StorageGetter, logger *zap.SugaredLogger) {
	writer.Header().Set("Content-Type", "application/json")
	data, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("get metric json, read request body error: %w", err)
		return
	}

	value, err := storage.GetMetricJSON(request.Context(), data)
	if err != nil {
		if value != nil {
			writer.WriteHeader(http.StatusNotFound)
		} else {
			writer.WriteHeader(http.StatusBadRequest)
		}
		logger.Warnf("get metric json error: %w", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	_, err = writer.Write(value)
	if err != nil {
		logger.Warnf("get metric json, write data to client error: %w", err)
	}
}

// проверка подключения к БД
func Ping(writer http.ResponseWriter, request *http.Request, storage StorageDB, logger *zap.SugaredLogger) {
	writer.Header().Set("Content-Type", "")
	err := storage.PingDB(request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Warnln(err)
		return
	}
	logger.Debug("database ping success")
	writer.WriteHeader(http.StatusOK)
}

// очистка storage
func Clear(writer http.ResponseWriter, request *http.Request, storage StorageDB, logger *zap.SugaredLogger) {
	writer.Header().Set("Content-Type", "")
	err := storage.Clear(request.Context())
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		logger.Warnln(err)
		return
	}
	logger.Debug("storage clear success")
	writer.WriteHeader(http.StatusOK)
}

// обновление списком json
func UpdateJSONSLice(writer http.ResponseWriter, request *http.Request, storage StorageSetter, logger *zap.SugaredLogger) {
	writer.Header().Set("Content-Type", "text/html")
	data, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("read request body error: %w", err)
		return
	}

	logger.Debugln("update metrics by json list", string(data))

	value, err := storage.UpdateJSONSlice(request.Context(), data)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		logger.Warnf("update metrics list error: %w", err)
		return
	}

	writer.WriteHeader(http.StatusOK)
	logger.Debugln("update metrics by json list success", string(value))
	_, err = writer.Write(value)
	if err != nil {
		logger.Warnf("write data to client error: %w", err)
	}
}
