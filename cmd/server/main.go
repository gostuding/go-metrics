package main

import (
	"fmt"
	"net/http"
)

// Создаём объект для хранения метрик
var metricStorage = NewMemStorage()

func main() {
	myServerMux := http.NewServeMux()
	myServerMux.HandleFunc("/", pathNotFound)
	myServerMux.HandleFunc("/update/", update)

	err := http.ListenAndServe(`:8080`, myServerMux)
	if err != nil {
		panic(err)
	}
}

// Заглушка для остальных запросов (не /update/...). Возвращает StatusBadGateway для всех запросов
func pathNotFound(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(http.StatusNotFound)
	fmt.Printf("Bad geteway: '%s', path: '%s'\r\n", request.Method, request.URL.Path)
}

// Обработка запроса на добавление или изменение метрики
func update(writer http.ResponseWriter, request *http.Request) {
	if request.Method != http.MethodPost {
		// разрешаем только POST-запросы
		fmt.Printf("Method not allowed: method: '%s', path: '%s'\r\n", request.Method, request.URL.Path)
		writer.WriteHeader(http.StatusMethodNotAllowed)
		writer.Write([]byte("Method not allowed. User POST method instead"))
		return
	}
	UpdateMemStorage(metricStorage, request.URL.Path, writer)
}

// Обновление данных у объекта, который относится к интерфейсу Storager, т.к. используется
// функция добавления и вывода в консоль (AddMetric и String)
func UpdateMemStorage(storage Storager, path string, writer http.ResponseWriter) {
	status, err := storage.AddMetric(path)
	writer.WriteHeader(status) // запись статуса для возврата
	if err != nil {
		fmt.Println(err)
	} else {
		fmt.Println(storage) // выводим в консоль изменённый объект memStorage
	}
}
