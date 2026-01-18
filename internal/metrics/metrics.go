package metrics

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type SensorsMetrics struct {
	temperature *prometheus.GaugeVec
	humidity    *prometheus.GaugeVec
	pressure    *prometheus.GaugeVec
}

func NewSensorsMetrics() *SensorsMetrics {
	return &SensorsMetrics{
		temperature: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sensor_temperature",
				Help: "Current temperature.",
			},
			[]string{"sensor_id", "unit"},
		),
		humidity: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sensor_humidity",
				Help: "Current humidity.",
			},
			[]string{"sensor_id", "unit"},
		),
		pressure: promauto.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: "sensor_pressure",
				Help: "Current pressure.",
			},
			[]string{"sensor_id", "unit"},
		),
	}
}

func (s *SensorsMetrics) RecordTemperature(_ context.Context, sensorID string, num float64) {
	s.temperature.WithLabelValues(sensorID, "celsius").Set(num)
}

func (s *SensorsMetrics) RecordHumidity(_ context.Context, sensorID string, num float64) {
	s.humidity.WithLabelValues(sensorID, "percent").Set(num)
}

func (s *SensorsMetrics) RecordPressure(_ context.Context, sensorID string, num float64) {
	s.pressure.WithLabelValues(sensorID, "hPa").Set(num)
}
