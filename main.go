package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"

	"go-analytics-service/internal/analytics"
	"go-analytics-service/internal/api"
	"go-analytics-service/internal/monitoring"
	"go-analytics-service/internal/storage"
)

func main() {
	analyzer := analytics.NewAnalyzer(50)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}
	store := storage.NewRedisStore(redisAddr)

	handler := api.NewHandler(analyzer, 1000, store)

	r := mux.NewRouter()

	// endpoint для метрик Prometheus
	r.Handle("/metrics", monitoring.Handler()).Methods(http.MethodGet)

	// наши основные API
	r.HandleFunc("/ingest", handler.IngestHandler).Methods(http.MethodPost)
	r.HandleFunc("/stats", handler.StatsHandler).Methods(http.MethodGet)
	r.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}).Methods(http.MethodGet)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Starting analytics service on :%s\n", port)

	// оборачиваем роутер в Prometheus middleware
	wrapped := monitoring.Middleware(r)

	if err := http.ListenAndServe(":"+port, wrapped); err != nil {
		log.Fatal(err)
	}
}
