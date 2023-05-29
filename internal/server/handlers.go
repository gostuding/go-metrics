package server

import (
	"fmt"
	"io"
	"net/http"
)

// -----------------------------------------------------------------------------------
// Определение интерфейсов
// -----------------------------------------------------------------------------------
// Интерфей для установки значений в объект из строки
type StorageSetter interface {
	Update(string, string, string) error
	UpdateJSON([]byte) ([]byte, error)
	Save() error
}

// Интерфейс получения значения метрики
type StorageGetter interface {
	GetMetric(string, string) (string, error)
	GetMetricJSON([]byte) ([]byte, error)
}

// Интерфейс для вывод значений в виде HTML
type HTMLGetter interface {
	GetMetricsHTML() string
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
func Update(writer http.ResponseWriter, request *http.Request, storage StorageSetter, metric updateMetricsArgs) {
	if err := storage.Update(metric.base.mType, metric.base.mName, metric.mValue); err != nil {
		writer.WriteHeader(http.StatusBadRequest)
	} else {
		writer.WriteHeader(http.StatusOK)
		saveStorage(storage)
	}
}

// Обработка запроса значения метрики
func GetMetric(writer http.ResponseWriter, request *http.Request, storage StorageGetter, metric getMetricsArgs) {
	value, err := storage.GetMetric(metric.mType, metric.mName)
	if err == nil {
		writer.WriteHeader(http.StatusOK)
		_, err = writer.Write([]byte(value))
	} else {
		writer.WriteHeader(http.StatusNotFound)
	}
	if err != nil {
		Logger.Warnf("write data to client error: %w", err)
	}
}

// Запрос всех метрик в html
func GetAllMetrics(writer http.ResponseWriter, request *http.Request, storage HTMLGetter) {
	writer.Header().Set("Content-Type", "text/html")
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte(storage.GetMetricsHTML()))
	if err != nil {
		Logger.Warnf("write metrics data to client error: %w", err)
	}
}

// обновление в JSON формате
func UpdateJSON(writer http.ResponseWriter, request *http.Request, storage StorageSetter) {
	writer.Header().Set("Content-Type", "application/json")
	data, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		Logger.Warnf("read request body error: %w", err)
	} else {
		value, err := storage.UpdateJSON(data)
		if err != nil {
			writer.WriteHeader(http.StatusBadRequest)
			Logger.Warnf("update metric error: %w", err)
		} else {
			writer.WriteHeader(http.StatusOK)
			saveStorage(storage)
			_, err = writer.Write(value)
			if err != nil {
				Logger.Warnf("write data to clie`nt error: %w", err)
			}
		}
	}
}

// получение метрики в JSON формате
func GetMetricJSON(writer http.ResponseWriter, request *http.Request, storage StorageGetter) {
	writer.Header().Set("Content-Type", "application/json")
	data, err := io.ReadAll(request.Body)
	if err != nil {
		writer.WriteHeader(http.StatusBadRequest)
		Logger.Warnf("read request body error: %w", err)
	} else {
		value, err := storage.GetMetricJSON(data)
		if err != nil {
			if value != nil {
				writer.WriteHeader(http.StatusNotFound)
			} else {
				writer.WriteHeader(http.StatusBadRequest)
			}
			_, err = writer.Write([]byte(fmt.Sprintf(`{"error": "%s"}`, err.Error())))
		} else {
			writer.WriteHeader(http.StatusOK)
			_, err = writer.Write(value)
		}
		if err != nil {
			Logger.Warnf("write data to client error: %w", err)
		}
	}
}

// сохранение хранилища
func saveStorage(storage StorageSetter) {
	if currentOptions.StoreInterval == 0 {
		err := storage.Save()
		if err != nil {
			Logger.Warnf("save storage error: %w", err)
		}
	}
}
