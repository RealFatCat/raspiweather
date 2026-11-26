package server

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/realfatcat/raspiweather/pkg/types"
)

const (
	// Graceful shutdown timeout
	shutdownTimeout = 5 * time.Second
	// Request timeouts
	readTimeout       = 5 * time.Second  // Maximum duration for reading the entire request
	readHeaderTimeout = 1 * time.Second  // Maximum duration for reading request headers
	writeTimeout      = 5 * time.Second  // Maximum duration before timing out writes of the response
	idleTimeout       = 60 * time.Second // Maximum amount of time to wait for the next request when keep-alives are enabled
)

type HTTP struct {
	srv              *http.Server
	producer         Producer
	backlightToggler BacklightToggler
}

func NewHTTP(address string, producer Producer, toggler BacklightToggler) *HTTP {
	h := &HTTP{
		producer:         producer,
		backlightToggler: toggler,
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/sensor-data", h.handleGetWeatherData)
	mux.HandleFunc("/toggle-leds", h.handleToggleLeds)

	h.srv = &http.Server{
		Addr:              address,
		Handler:           mux,
		ReadTimeout:       readTimeout,
		ReadHeaderTimeout: readHeaderTimeout,
		WriteTimeout:      writeTimeout,
		IdleTimeout:       idleTimeout,
	}
	return h
}

// sensorDataResponse is the JSON body returned by /sensor-data.
// Data holds the readings that succeeded.
// Errors describes sensors that failed.
// A partial response (some data, some errors) is still a 200 OK.
type sensorDataResponse struct {
	Data   []types.WeatherData   `json:"data"`
	Errors []sensorErrorResponse `json:"errors,omitempty"`
}

type sensorErrorResponse struct {
	SensorID string `json:"sensor_id"`
	Error    string `json:"error"`
}

func (h *HTTP) handleGetWeatherData(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	res := h.producer.ReadWeatherData()

	resp := sensorDataResponse{Data: res.Data}
	for _, e := range res.Errors {
		msg := "unknown error"
		if e.Err != nil {
			msg = e.Err.Error()
		}
		resp.Errors = append(resp.Errors, sensorErrorResponse{SensorID: e.SensorID, Error: msg})
	}

	w.Header().Set("Content-Type", "application/json")
	if len(res.Data) == 0 {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
	}

	if err := json.NewEncoder(w).Encode(resp); err != nil {
		slog.Error("encoding sensor data", "error", err)
	}
}

func (h *HTTP) handleToggleLeds(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if h.backlightToggler == nil {
		errMsg := "toggler is not set"
		slog.Error(errMsg)
		http.Error(w, errMsg, http.StatusInternalServerError)
		return
	}

	if err := h.backlightToggler.ToggleBacklight(); err != nil {
		slog.Error("toggling backlight", "error", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(map[string]string{"status": "ok"}); err != nil {
		slog.Error("encoding toggle-leds response", "error", err)
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

		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()

		if err := h.srv.Shutdown(shutdownCtx); err != nil {
			slog.Error("Server shutdown error", "error", err)
			return err
		}
		slog.Info("Server stopped")
		return nil
	}
}
