package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/gostuding/go-metrics/cmd/server/handlers"
	"github.com/gostuding/go-metrics/cmd/server/storage"
)

var memory = storage.MemStorage{}

func GetRouter() http.Handler {
	router := chi.NewRouter()

	router.Use(middleware.RealIP)
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)

	router.Get("/", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetAllMetrics(w, r, memory)
	})
	router.Post("/update/{mType}/{mName}/{mValue}", func(w http.ResponseWriter, r *http.Request) {
		handlers.Update(w, r, &memory)
	})
	router.Get("/value/{mType}/{mName}", func(w http.ResponseWriter, r *http.Request) {
		handlers.GetMetric(w, r, memory)
	})

	return router
}
