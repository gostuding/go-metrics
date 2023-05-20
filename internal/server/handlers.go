package server

import (
	"net/http"
)

// -----------------------------------------------------------------------------------
// Определение интерфейсов
// -----------------------------------------------------------------------------------
// Интерфей для установки значений в объект из строки
type StorageSetter interface {
	Update(string, string, string) error
}

// Интерфейс получения значения метрики
type StorageGetter interface {
	GetMetric(string, string) (string, error)
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
		_, err = writer.Write([]byte(err.Error()))
	}
	if err != nil {
		Logger.Warnf("write data to client error: %v", err)
	}
}

// Запрос всех метрик в html
func GetAllMetrics(writer http.ResponseWriter, request *http.Request, storage HTMLGetter) {
	writer.WriteHeader(http.StatusOK)
	_, err := writer.Write([]byte(storage.GetMetricsHTML()))
	if err != nil {
		Logger.Warnf("write metrics data to client error: %v", err)
	}
}
