package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// интерфейс для обработки запросов
type Storage interface {
	StorageSeter
	StorageGetter
	HTMLGetter
}

// получние однотипных данных из адреса запроса
func getUrlParams(r *http.Request) metricsArgs {
	return metricsArgs{mType: chi.URLParam(r, "mType"), mName: chi.URLParam(r, "mName"), mValue: chi.URLParam(r, "mValue")}
}

// формирование доступных адресов
func makeRouter(storage Storage) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		GetAllMetrics(w, r, storage)
	})
	router.Post("/update/{mType}/{mName}/{mValue}", func(w http.ResponseWriter, r *http.Request) {
		Update(w, r, storage, getUrlParams(r))
	})
	router.Get("/value/{mType}/{mName}", func(w http.ResponseWriter, r *http.Request) {
		GetMetric(w, r, storage, getUrlParams(r))
	})

	return router
}
