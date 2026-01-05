package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

const (
	// Graceful shutdown timeout
	shutdownTimeout = 5 * time.Second
	// Request timeouts
	readTimeout  = 5 * time.Second  // Maximum duration for reading the entire request
	writeTimeout = 5 * time.Second  // Maximum duration before timing out writes of the response
	idleTimeout  = 60 * time.Second // Maximum amount of time to wait for the next request when keep-alives are enabled
)

type HTTP struct {
	srv      *http.Server
	producer Producer
}

func NewHTTP(address string, producer Producer) *HTTP {
	h := &HTTP{
		producer: producer,
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/sensor-data", h.handleGetWeatherData)

	h.srv = &http.Server{
		Addr:         address,
		Handler:      mux,
		ReadTimeout:  readTimeout,
		WriteTimeout: writeTimeout,
		IdleTimeout:  idleTimeout,
	}
	return h
}

func (h *HTTP) handleGetWeatherData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	data, err := h.producer.ReadWeatherData()
	if err != nil {
		slog.Error("reading sensor data", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	if err := json.NewEncoder(w).Encode(data); err != nil {
		slog.Error("encoding sensor data", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (h *HTTP) ListenAndServe(ctx context.Context) error {
	serverErr := make(chan error, 1)
	go func() {
		slog.Info("Starting HTTP Server", "address", h.srv.Addr)
		slog.Info("Metrics would be available at /metrics")
		if err := h.srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("HTTP Server error", "error", err)
			serverErr <- err
		}
	}()

	select {
	case err := <-serverErr:
		return err
	case <-ctx.Done():
		slog.Info("Shutting down server...")

		shutdownCtx, cancel := context.WithTimeout(ctx, shutdownTimeout)
		defer cancel()

		if err := h.srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("Server shutdown error", "error", err)
			return err
		}
		slog.Info("Server stopped")
		return nil
	}
}
