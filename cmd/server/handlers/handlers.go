package handlers

import (
	"fmt"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/gostuding/go-metrics/cmd/server/storage"
)

// Заглушка для остальных запросов (не /update/...). Возвращает StatusBadGateway для всех запросов
func PathNotFound(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	fmt.Printf("Bad geteway: '%s', path: '%s'\r\n", request.Method, request.URL.Path)
}

func MethosNotAllowed(writer http.ResponseWriter, request *http.Request) {
	fmt.Printf("Method not allowed: method: '%s', path: '%s'\r\n", request.Method, request.URL.Path)
	writer.WriteHeader(http.StatusMethodNotAllowed)
	writer.Write([]byte("Method not allowed. User POST method instead"))
}

// Обработка запроса на добавление или изменение метрики
func Update(writer http.ResponseWriter, request *http.Request, storage storage.StorageSeter) {

	status, err := storage.Update(chi.URLParam(request, "mType"),
		chi.URLParam(request, "mName"), chi.URLParam(request, "mValue"))
	writer.WriteHeader(status) // запись статуса для возврата
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(storage) // выводим в консоль изменённый объект memStorage
	}
}

// Обработка запроса метрики
func GetMetric(writer http.ResponseWriter, request *http.Request, storage storage.StorageGetter) {
	value, status := storage.GetMetric(chi.URLParam(request, "mType"), chi.URLParam(request, "mName"))
	writer.WriteHeader(status)
	writer.Write([]byte(value))
}

// Запрос всех метрик в html
func GetAllMetrics(writer http.ResponseWriter, request *http.Request, storage storage.HtmlGetter) {
	writer.WriteHeader(http.StatusOK)
	writer.Write([]byte(storage.GetMetricsHTML()))
}
