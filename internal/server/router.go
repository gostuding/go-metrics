package server

import (
	"io"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"
)

// private func.
func updateParams(r *http.Request) updateMetricsArgs {
	return updateMetricsArgs{
		base:   getMetricsArgs{mType: chi.URLParam(r, "mType"), mName: chi.URLParam(r, "mName")},
		mValue: chi.URLParam(r, "mValue"),
	}
}

// private func.
func getParams(r *http.Request) getMetricsArgs {
	return getMetricsArgs{mType: chi.URLParam(r, "mType"), mName: chi.URLParam(r, "mName")}
}

// private func. Create posible hendlers for server.
func makeRouter(storage Storage, logger *zap.SugaredLogger, key []byte) http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP, hashCheckMiddleware(key, logger, false), gzipMiddleware(logger),
		loggerMiddleware(logger), middleware.Recoverer)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := GetAllMetrics(r.Context(), storage)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Warnf("get all metrics error: %w", w)
		} else {
			w.Header().Set("Content-Type", "text/html")
			_, err = w.Write([]byte(body))
			if err != nil {
				logger.Warnf("write metrics data to client error: %w", err)
			}
		}
	})

	router.Post("/value/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Warnf("get metric json, read request body error: %w", err)
			return
		}
		body, status, err := GetMetricJSON(r.Context(), storage, body)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		if err != nil {
			logger.Warnf("get metric by json error: %w", err)
		}
		_, err = w.Write([]byte(body))
		if err != nil {
			logger.Warnf("write data to client error: %w", err)
		}
	})

	router.Get("/value/{mType}/{mName}", func(w http.ResponseWriter, r *http.Request) {
		body, err := GetMetric(r.Context(), storage, getParams(r))
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			logger.Warnf("get metric error: %w", err)
		} else {
			_, err = w.Write([]byte(body))
			if err != nil {
				logger.Warnf("write data to client error: %w", err)
			}
		}
	})

	router.Post("/update/{mType}/{mName}/{mValue}", func(w http.ResponseWriter, r *http.Request) {
		m := updateParams(r)
		status, err := Update(r.Context(), storage, m)
		w.WriteHeader(status)
		if err != nil {
			logger.Warnf(err.Error())
		} else {
			logger.Debugf("update metric '%s' success", m.base.mType)
		}
	})

	router.Post("/update/", func(w http.ResponseWriter, r *http.Request) {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Warnf("read request body error: %w", err)
			return
		}
		data, err := UpdateJSON(r.Context(), body, storage)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Warnf("update metric request error: %w", err)
		} else {
			logger.Debug("update metric by json success")
			w.Header().Set("Content-Type", "application/json")
			_, err = w.Write(data)
			if err != nil {
				logger.Warnf("write data to client error: %w", err)
			}
		}
	})

	router.Post("/updates/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Warnf("read request body error: %w", err)
			return
		}
		data, err := UpdateJSONSLice(r.Context(), body, storage)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Warnf("update metrics by slice error: %w", err)
		} else {
			logger.Debug("update metrics by json list success")
			_, err = w.Write(data)
			if err != nil {
				logger.Warnf("write data to client error: %w", err)
			}
		}

	})

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		status, err := Ping(r.Context(), storage)
		w.WriteHeader(status)
		w.Header().Set("Content-Type", "")
		if err != nil {
			logger.Warnf(err.Error())
		} else {
			logger.Debug("database ping success")
		}

	})

	router.Get("/clear", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "")
		status, err := Clear(r.Context(), storage)
		w.WriteHeader(status)
		if err != nil {
			logger.Warnf("clear request error: %w", err)
		}
	})

	router.Mount("/debug", middleware.Profiler())

	return router
}
