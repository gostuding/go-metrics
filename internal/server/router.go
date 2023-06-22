package server

import (
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// интерфейс для обработки запросов
type Storage interface {
	StorageSetter
	StorageGetter
	HTMLGetter
	StorageDB
}

// получние однотипных данных из адреса запроса
func updateParams(r *http.Request) updateMetricsArgs {
	return updateMetricsArgs{base: getMetricsArgs{mType: chi.URLParam(r, "mType"), mName: chi.URLParam(r, "mName")}, mValue: chi.URLParam(r, "mValue")}
}
func getParams(r *http.Request) getMetricsArgs {
	return getMetricsArgs{mType: chi.URLParam(r, "mType"), mName: chi.URLParam(r, "mName")}
}

// формирование доступных адресов
func makeRouter(storage Storage, logger *zap.SugaredLogger, key string) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP, hashCheckMiddleware(key, logger), gzipMiddleware(logger), loggerMiddleware(logger), middleware.Recoverer)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		GetAllMetrics(w, r, storage, logger, key)
	})
	router.Post("/value/", func(w http.ResponseWriter, r *http.Request) {
		GetMetricJSON(w, r, storage, logger, key)
	})
	router.Get("/value/{mType}/{mName}", func(w http.ResponseWriter, r *http.Request) {
		GetMetric(w, r, storage, getParams(r), logger, key)
	})
	router.Post("/update/{mType}/{mName}/{mValue}", func(w http.ResponseWriter, r *http.Request) {
		Update(w, r, storage, updateParams(r), logger)
	})
	router.Post("/update/", func(w http.ResponseWriter, r *http.Request) {
		UpdateJSON(w, r, storage, logger, key)
	})
	router.Post("/updates/", func(w http.ResponseWriter, r *http.Request) {
		UpdateJSONSLice(w, r, storage, logger, key)
	})
	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		Ping(w, r, storage, logger)
	})
	router.Get("/clear", func(w http.ResponseWriter, r *http.Request) {
		Clear(w, r, storage, logger)
	})

	return router
}
