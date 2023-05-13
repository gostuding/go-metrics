package server

import (
	"log"
	"net/http"
)

// -----------------------------------------------------------------------------------
// Определение интерфейсов
// -----------------------------------------------------------------------------------
// Интерфей для установки значений в объект из строки
type StorageSeter interface {
	Update(string, string, string) (int, error)
}

// Интерфейс получения значения метрики
type StorageGetter interface {
	GetMetric(string, string) (string, int)
}

// Интерфейс для вывод значений в виде HTML
type HTMLGetter interface {
	GetMetricsHTML() string
}

// -----------------------------------------------------------------------------------
// Определение функций, которые используют интерфейсы
// -----------------------------------------------------------------------------------
// дабы не раздувать количество аргументов
// ввел структуру для передачи в функции
// также это позволяет не завязываться на обработчик, который используется при роутинге
type metricsArgs struct {
	mType  string
	mName  string
	mValue string
}

// Обработка запроса на добавление или изменение метрики
func Update(writer http.ResponseWriter, request *http.Request, storage StorageSeter, metric metricsArgs) {
	status, err := storage.Update(metric.mType, metric.mName, metric.mValue)
	writer.WriteHeader(status)
	// логирование ошибки
	if err != nil {
		log.Printf("update metric error: %v", err)
	}
}

// Обработка запроса значения метрики
func GetMetric(writer http.ResponseWriter, request *http.Request, storage StorageGetter, metric metricsArgs) {
	value, status := storage.GetMetric(metric.mType, metric.mName)
	writer.WriteHeader(status)
	writer.Write([]byte(value))
}

// Запрос всех метрик в html
func GetAllMetrics(writer http.ResponseWriter, request *http.Request, storage HTMLGetter) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(storage.GetMetricsHTML()))
}
