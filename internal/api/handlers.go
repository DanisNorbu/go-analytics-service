package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"go-analytics-service/internal/analytics"
	"go-analytics-service/internal/monitoring"
	"go-analytics-service/internal/storage"
)

type Handler struct {
	Analyzer *analytics.Analyzer
	InputCh  chan analytics.Metric
	Store    *storage.RedisStore
}

func NewHandler(a *analytics.Analyzer, bufferSize int, store *storage.RedisStore) *Handler {
	if bufferSize <= 0 {
		bufferSize = 1000
	}

	h := &Handler{
		Analyzer: a,
		InputCh:  make(chan analytics.Metric, bufferSize),
		Store:    store,
	}

	go func() {
		for m := range h.InputCh {
			res := h.Analyzer.AddMetric(m)

			monitoring.MetricsProcessedTotal.Inc()

			if res.IsAnomalyRPS {
				monitoring.AnomaliesTotal.Inc()

				log.Printf("[ANOMALY] ts=%s rps=%.2f z=%.2f avg_rps=%.2f\n",
					res.LastTimestamp.Format(time.RFC3339),
					res.LastRPS, res.ZScoreRPS, res.AvgRPS)
			}

			if h.Store != nil {
				h.Store.SaveMetric(m)
			}
		}
	}()
	return h
}

type MetricInput struct {
	Timestamp int64   `json:"timestamp"`
	CPU       float64 `json:"cpu"`
	RPS       float64 `json:"rps"`
}

func (h *Handler) IngestHandler(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()
	var in MetricInput
	if err := json.NewDecoder(r.Body).Decode(&in); err != nil {
		http.Error(w, "invalid JSON: "+err.Error(), http.StatusBadRequest)
		return
	}

	ts := time.Now()
	if in.Timestamp != 0 {
		ts = time.UnixMilli(in.Timestamp)
	}

	m := analytics.Metric{
		Timestamp: ts,
		CPU:       in.CPU,
		RPS:       in.RPS,
	}

	select {
	case h.InputCh <- m:
		w.WriteHeader(http.StatusAccepted)
		_, _ = w.Write([]byte(`{"status":"accepted"}`))
	default:
		http.Error(w, "ingest queue full", http.StatusServiceUnavailable)
	}
}

func (h *Handler) StatsHandler(w http.ResponseWriter, r *http.Request) {
	res := h.Analyzer.GetLastResult()
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(res); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
