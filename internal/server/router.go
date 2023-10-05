package server

import (
	"crypto/rsa"
	"io"
	"net"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
	"go.uber.org/zap"

	"github.com/gostuding/go-metrics/internal/server/middlewares"
)

// Internal constants.
const (
	writeErrorString = "write data to client error: %w"
	mTypeString      = "mType"
	mNameString      = "mName"

	contentEncoding = "Content-Encoding"
	contentType     = "Content-Type"
	gzipString      = "gzip"
	applicationJSON = "application/json"
	textHTML        = "text/html"
	hashVarName     = "HashSHA256"
)

// private func. Create posible hendlers for server.
func makeRouter(
	storage Storage,
	logger *zap.SugaredLogger,
	hashKey []byte,
	pk *rsa.PrivateKey,
	subnet *net.IPNet,
) http.Handler {
	router := chi.NewRouter()
	router.Use(
		middleware.RealIP,
		middlewares.SubNetCheckMiddleware(subnet, logger),
		middlewares.HashCheckMiddleware(hashKey, logger),
		middlewares.GzipMiddleware(logger),
		middlewares.DecriptMiddleware(pk, logger),
		middlewares.LoggerMiddleware(logger),
		middleware.Recoverer,
	)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		body, err := GetAllMetrics(r.Context(), storage)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Warnf("get all metrics error: %w", w)
		} else {
			w.Header().Set(contentType, textHTML)
			_, err = w.Write(body)
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
		w.Header().Set(contentType, applicationJSON)
		w.WriteHeader(status)
		if err != nil {
			logger.Warnf("get metric by json error: %w", err)
		}
		_, err = w.Write(body)
		if err != nil {
			logger.Warnf(writeErrorString, err)
		}
	})

	router.Get("/value/{mType}/{mName}", func(w http.ResponseWriter, r *http.Request) {
		body, err := GetMetric(
			r.Context(),
			storage,
			getMetricsArgs{
				mType: chi.URLParam(r, mTypeString),
				mName: chi.URLParam(r, mNameString),
			},
		)
		if err != nil {
			w.WriteHeader(http.StatusNotFound)
			logger.Warnf("get metric error: %w", err)
		} else {
			_, err = w.Write(body)
			if err != nil {
				logger.Warnf(writeErrorString, err)
			}
		}
	})

	router.Post("/update/{mType}/{mName}/{mValue}", func(w http.ResponseWriter, r *http.Request) {
		m := updateMetricsArgs{
			base: getMetricsArgs{
				mType: chi.URLParam(r, mTypeString),
				mName: chi.URLParam(r, mNameString),
			},
			mValue: chi.URLParam(r, "mValue"),
		}
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
			logger.Warnf("update read request body error: %w", err)
			return
		}
		data, err := UpdateJSON(r.Context(), body, storage)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Warnf("update metric request error: %w", err)
		} else {
			logger.Debug("update metric by json success")
			w.Header().Set(contentType, applicationJSON)
			_, err = w.Write(data)
			if err != nil {
				logger.Warnf(writeErrorString, err)
			}
		}
	})

	router.Post("/updates/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, textHTML)
		body, err := io.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Warnf("updates read request body error: %w", err)
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
				logger.Warnf(writeErrorString, err)
			}
		}
	})

	router.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
		status, err := Ping(r.Context(), storage)
		w.WriteHeader(status)
		w.Header().Set(contentType, "")
		if err != nil {
			logger.Warnf(err.Error())
		} else {
			logger.Debug("database ping success")
		}
	})

	router.Get("/clear", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set(contentType, "")
		status, err := Clear(r.Context(), storage)
		w.WriteHeader(status)
		if err != nil {
			logger.Warnf("clear request error: %w", err)
		}
	})

	router.Mount("/debug", middleware.Profiler())
	return router
}
