package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

// интерфейс для обработки запросов
type Storage interface {
	StorageSetter
	StorageGetter
	HTMLGetter
}

// получние однотипных данных из адреса запроса
func updateParams(r *http.Request) updateMetricsArgs {
	return updateMetricsArgs{base: getMetricsArgs{mType: chi.URLParam(r, "mType"), mName: chi.URLParam(r, "mName")}, mValue: chi.URLParam(r, "mValue")}
}
func getParams(r *http.Request) getMetricsArgs {
	return getMetricsArgs{mType: chi.URLParam(r, "mType"), mName: chi.URLParam(r, "mName")}
}

// формирование доступных адресов
func makeRouter(storage Storage) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(serverMiddleware)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		GetAllMetrics(w, r, storage)
	})
	router.Post("/update/{mType}/{mName}/{mValue}", func(w http.ResponseWriter, r *http.Request) {
		Update(w, r, storage, updateParams(r))
	})
	router.Get("/value/{mType}/{mName}", func(w http.ResponseWriter, r *http.Request) {
		GetMetric(w, r, storage, getParams(r))
	})

	return router
}
